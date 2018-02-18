package vpn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Event that is emitted when client's bandwidth usage has been changed.
// For more details, please see the OpenVPN docs at:
// https://openvpn.net/index.php/open-source/documentation/miscellaneous/79-management-interface.html
type ServerEventByteCount struct {
	ClientSessionID uint64
	BytesIn         uint64
	BytesOut        uint64
}

// ---------------------------------------------------------------------------------------------------------------------

// Abstract interface for the remote VPN management interface.
// At this moment is implemented only for Linux, but also should be implemented for other OS too.
type TelnetCommunicator interface {
	Connect() error
	Disconnect()

	KillUserSession(userCommonName string) error

	SubscribeForByteCountEvents(timeoutSeconds uint8) (<-chan *ServerEventByteCount, error)

	// todo: think about adding server disconnection events
	// for appropriate disconnects handling
}

// ---------------------------------------------------------------------------------------------------------------------

// Internal events types, used in Linux implementation of TelnetCommunicator.
// todo: move this into it's own package when source base would be under /src
const (
	kEventTypeConnectionStateChanged = 0
	kEventTypeUserSessionKilled      = 1
	kEventTypeByteCount              = 2
)

// Base interface for all internal events used by linux implementation of TelnetCommunicator.
type internalEvent interface {
	typeID() uint16
}

// ---------------------------------------------------------------------------------------------------------------------

// todo: add appropriate disconnect handling
type internalEventConnectionStateChanged struct {
	IsConnected bool
}

func (e *internalEventConnectionStateChanged) typeID() uint16 {
	return kEventTypeConnectionStateChanged
}

// ---------------------------------------------------------------------------------------------------------------------

type internalEventClientKilled struct {
	CommonName string
}

func (e *internalEventClientKilled) typeID() uint16 {
	return kEventTypeUserSessionKilled
}

// ---------------------------------------------------------------------------------------------------------------------

type internalEventByteCount struct {
	ClientID uint64
	BytesIn  uint64
	BytesOut uint64
}

func (e *internalEventByteCount) typeID() uint16 {
	return kEventTypeByteCount
}

// ---------------------------------------------------------------------------------------------------------------------

const (
	kCommandExecutionTimeout = time.Second * 5
	kEventTransferTimeout    = time.Millisecond * 100
)

// Linux implementation of TelnetCommunicator.
// Uses standard telnet bash command in sub-process for communicating with remote management interface.
// todo: add telnet to the project dependencies.
type LinuxTelnetCommunicator struct {
	Host string
	Port uint16

	// Used for controlling telnet command sub-process.
	processHandler *exec.Cmd
	processInPipe  io.WriteCloser
	processOutPipe io.ReadCloser

	externalByteCountEvents chan *ServerEventByteCount

	// Internal events.
	// Used for reacting on server side notifications.

	// There is non-zero probability, that events would arrive very quickly.
	// To prevent concurrent maps access (foe example, clientsKilledEventsRegistry),
	// and panics as a result - mutex is used for synchronization purposes.
	internalEventsMutex sync.Mutex

	// Events about connection/disconnection and bandwidth usage are not callers-relevant.
	// So simple common channel is enough.
	connectionEvents       chan *internalEventConnectionStateChanged
	clientsByteCountEvents chan *internalEventByteCount

	// Events about clients killed are name-relevant.
	// There is a non-zero probability of parallel occurrence of several such event at a time.
	// To avoid calling goroutines to be confused - simple events dispatching mechanics is used.
	//
	// Each one call to KillUserSession() binds new channel, in which the event itself will arrive.
	clientsKilledEventsRegistry map[string]chan *internalEventClientKilled

	isClosed bool
}

func NewLinuxTelnetCommunicator(host string, port uint16) *LinuxTelnetCommunicator {
	return &LinuxTelnetCommunicator{
		Host: host,
		Port: port,

		processHandler: nil,

		externalByteCountEvents: make(chan *ServerEventByteCount),

		connectionEvents:       make(chan *internalEventConnectionStateChanged),
		clientsByteCountEvents: make(chan *internalEventByteCount),

		clientsKilledEventsRegistry: make(map[string]chan *internalEventClientKilled),

		isClosed: false,
	}
}

// Attempts to connect to the remote control interface.
// Blocks for kCommandExecutionTimeout in case of no response from the server.
// Returns error in case if no connection was established.
// todo: might not connect if there is already connected interface from other ttl. this must be fixed.
func (c *LinuxTelnetCommunicator) Connect() error {
	if c.isClosed {
		return errors.New("communicator is now closed. please, use new one for the new connection")
	}

	c.processHandler = exec.Command("telnet", c.Host, fmt.Sprint(c.Port))

	var err error
	c.processInPipe, err = c.processHandler.StdinPipe()
	if err != nil {
		return errors.New("can't connect to sub-process standard input pipe")
	}

	c.processOutPipe, err = c.processHandler.StdoutPipe()
	if err != nil {
		return errors.New("can't connect to sub-process standard out pipe")
	}
	c.attachToStdOutAndProcessMessages()

	// Start the process and give it some time to establish the connection.
	err = c.processHandler.Start()
	if err != nil {
		return WrapError(err, "can't start sub-process")
	}

	select {
	case e := <-c.connectionEvents:
		{
			if !e.IsConnected {
				return errors.New("can't establish connection with remote management interface")

			} else {
				// Seems to be ok.
				return nil
			}
		}

	case <-time.After(kCommandExecutionTimeout):
		{
			return errors.New("can't establish connection with remote management interface")
		}
	}
}

func (c *LinuxTelnetCommunicator) Disconnect() {
	c.processHandler.Process.Kill()
	c.processInPipe.Close()
	c.processOutPipe.Close()
	c.isClosed = true
}

// Attempts to disconnect the user from the server.
// Blocks for kCommandExecutionTimeout in case of no response from the server.
func (c *LinuxTelnetCommunicator) KillUserSession(userCommonName string) error {
	// Accessing to the events registry should be synchronous.
	c.internalEventsMutex.Lock()

	// Checking of command exclusiveness.
	// In case if same command is already registered - no need for duplicates.
	_, isChannelAlreadyPresent := c.clientsKilledEventsRegistry[userCommonName]
	if isChannelAlreadyPresent {
		c.internalEventsMutex.Unlock()
		return errors.New("it seems, that other goroutine is waiting for the result of the same command")
	}

	// Creating channel for the incoming event.
	// Warn: channel MUST be created before any command would be transferred to the server.
	// otherwise - the incoming event might be missed.
	channel := make(chan *internalEventClientKilled)
	c.clientsKilledEventsRegistry[userCommonName] = channel

	// On method exit - channel must be removed along with map record.
	defer func() {
		c.internalEventsMutex.Lock()
		delete(c.clientsKilledEventsRegistry, userCommonName)
		c.internalEventsMutex.Unlock()
	}()

	// Unlocking is necessary to be done before the code reaches event checking stage.
	// Otherwise - the dispatch method would be unable to deliver the event.
	c.internalEventsMutex.Unlock()

	// Transferring the command itself and waiting for the response.
	command := fmt.Sprint("kill ", userCommonName, "\n")
	_, err := c.processInPipe.Write([]byte(command))
	if err != nil {
		return WrapError(err, "can't transfer command to the internal sub-process standard input pipe")
	}

	select {
	case <-channel:
		{
			// Ok, command executed successfully.
			return nil
		}

	case <-time.After(kCommandExecutionTimeout):
		{
			return errors.New(
				"can't transfer command to the remote management interface, " +
					"or there is no expected response")
		}
	}
}

// Subscribes for byte count events.
//
// "timeoutSeconds" specifies how often next event would be emitted.
// In case if 0 would be received as a "timeoutSeconds" argument value - it would be silently corrected to 1,
// because, by the OpenVPN specification - 0 is determined as "no events at all".
func (c *LinuxTelnetCommunicator) SubscribeForByteCountEvents(
	timeoutSeconds uint8) (<-chan *ServerEventByteCount, error) {

	if timeoutSeconds == 0 {
		timeoutSeconds = 1
	}

	// Transferring the command itself and waiting for the response.
	command := fmt.Sprint("bytecount ", fmt.Sprint(timeoutSeconds), "\n")
	_, err := c.processInPipe.Write([]byte(command))
	if err != nil {
		return nil, WrapError(err, "can't transfer command to the internal sub-process standard input pipe")
	}

	return c.externalByteCountEvents, nil
}

// Non-blocking.
// Reads internal sub-process std out and processes each one event occurred.
func (c *LinuxTelnetCommunicator) attachToStdOutAndProcessMessages() {
	go func() {
		reader := bufio.NewReader(c.processOutPipe)

		for {
			if c.isClosed {
				// Communicator was closed.
				// No any responses should be read.
				return
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				c.logError(err.Error())
				continue

				// todo: think to reconnect on error
			}

			c.processOutput(line)

			// There is no reason to drain the processor.
			time.Sleep(time.Millisecond * 10)
		}
	}()
}

// Parses line received from the internal sub-process and extracts some events.
// In case of successfully parsed one - reports it for the further processing.
// In case of error - writes to the log, because there is no any goroutines waiting for this one.
func (c *LinuxTelnetCommunicator) processOutput(line string) {

	// The byte count event seems to be most frequent.
	// So it is fine to check it first, because there is a very high probability of success here.
	if strings.Contains(line, "BYTECOUNT_CLI") {
		line = strings.SplitAfter(line, ":")[1]
		line = strings.Replace(line, "\n", "", -1)

		fields := strings.Split(line, ",")
		if len(fields) != 3 {
			c.logError("unexpected output received: " + line)
			return
		}

		clientID, _ := strconv.ParseUint(fields[0], 10, 64)
		bytesIn, _ := strconv.ParseUint(fields[1], 10, 64)
		bytesOut, _ := strconv.ParseUint(fields[2], 10, 64)

		event := &internalEventByteCount{
			ClientID: clientID,
			BytesIn:  bytesIn,
			BytesOut: bytesOut,
		}
		c.dispatchIncomingEvent(event)
		return
	}

	if strings.Contains(line, "client(s) killed") {
		if strings.Contains(line, "SUCCESS") {
			line = strings.Replace(line, "'", "", -1)

			words := strings.Fields(line)
			if len(words) < 4 {
				c.logError("unexpected output received: " + line)
				return
			}

			killedUserName := strings.Fields(line)[3]
			event := &internalEventClientKilled{CommonName: killedUserName}
			c.dispatchIncomingEvent(event)
			return
		}
	}

	// Connection event occurs only once or two per session.
	// So, for the performance purposes it should be checked in last order.
	if strings.Contains(line, "Connected to") {
		c.connectionEvents <- &internalEventConnectionStateChanged{IsConnected: true}
		return
	}
}

func (c *LinuxTelnetCommunicator) dispatchIncomingEvent(event internalEvent) {
	switch event.typeID() {
	case kEventTypeConnectionStateChanged:
		{
			select {
			case c.connectionEvents <- event.(*internalEventConnectionStateChanged):
				{
					// Ok, event transferred well.
				}

			case <-time.After(kEventTransferTimeout):
				{
					c.logError(
						"Connection state changed event can't be dispatched, " +
							"because it seems that no one goroutine expects it. " +
							"Event dropped to prevent whole events flow hanging.")
				}
			}
		}

	case kEventTypeByteCount:
		{
			internalEvent := event.(*internalEventByteCount)
			externalEvent := &ServerEventByteCount{
				ClientSessionID: internalEvent.ClientID,
				BytesIn:         internalEvent.BytesIn,
				BytesOut:        internalEvent.BytesOut,
			}

			select {
			case c.externalByteCountEvents <- externalEvent:
				{
					// Ok, event transferred well.
				}

			case <-time.After(kEventTransferTimeout):
				{
					c.logError(
						"Byte count event can't be dispatched, " +
							"because it seems that no one goroutine expects it. " +
							"Event dropped to prevent whole events flow hanging.")
				}
			}
		}

	default:
		{
			c.synchronouslyRegisterEvent(event)
		}
	}
}

func (c *LinuxTelnetCommunicator) synchronouslyRegisterEvent(event internalEvent) {
	c.internalEventsMutex.Lock()
	defer c.internalEventsMutex.Unlock()

	switch event.typeID() {
	case kEventTypeUserSessionKilled:
		{
			e := event.(*internalEventClientKilled)
			channel, isPresent := c.clientsKilledEventsRegistry[e.CommonName]
			if !isPresent {
				c.logError(
					"No channel found for user killed event. " +
						"Event dropped. " +
						"Target user name: " + e.CommonName)
				return
			}

			select {
			case channel <- e:
				{
					// Ok, event was transferred well.
				}

			case <-time.After(kEventTransferTimeout):
				{
					c.logError(
						"User killed event arrived, but can't be delivered for further processing," +
							"because it seems that no one goroutine is able to process it. " +
							"Event dropped. " +
							"Target user name: " + e.CommonName)
				}
			}
		}

	default:
		c.logError(
			"Unexpected event arrived. " +
				"Event type: " + fmt.Sprint(event.typeID()))
	}
}

func (c *LinuxTelnetCommunicator) logError(message string) {
	log.Println("ERROR: telnet management interface > " + message)
}

// ---------------------------------------------------------------------------------------------------------------------

// note: [suggestion] rename to ServerController.
type Control struct {
	VPNServerAddress     string
	ManagementTelnetPort uint16

	communicator TelnetCommunicator
}

func NewControl(VPNServerAddress string, ManagementTelnetPort uint16) (*Control, error) {
	controller := &Control{
		VPNServerAddress:     VPNServerAddress,
		ManagementTelnetPort: ManagementTelnetPort,

		// todo: [feature, cross-platform] add platform-specific communicator type select here.
		// By default, Linux-specific communicator is used for now.
		communicator: NewLinuxTelnetCommunicator(VPNServerAddress, ManagementTelnetPort),
	}

	// Attempt to establish connection with remote management interface.
	err := controller.communicator.Connect()
	if err != nil {
		return nil, WrapError(err, "can't connect to management interface")
	}

	return controller, nil
}

// Returns read only channel of events, the describes traffic usage by the specific client.
// Please, see "ServerEventByteCount" struct documentation for more details.
//
// "timeoutSeconds" specifies how often next event would be emitted.
// In case if 0 would be received as a "timeoutSeconds" argument value - it would be silently corrected to 1,
// because, by the OpenVPN specification - 0 is determined as "no events at all".
//
// Note: in case if client doesn't used any bandwidth - server event might be not generated.
func (c *Control) SubscribeForByteCountEvents(timeoutSeconds uint8) (<-chan *ServerEventByteCount, error) {
	return c.communicator.SubscribeForByteCountEvents(timeoutSeconds)
}

func (c *Control) Close() {
	c.communicator.Disconnect()

}

func (c *Control) KillUserSession(userCommonName string) error {
	return c.communicator.KillUserSession(userCommonName)
}

// ---------------------------------------------------------------------------------------------------------------------

// todo: move this into specific package
func WrapError(err error, message string) error {
	return errors.New(message + " > " + err.Error())
}
