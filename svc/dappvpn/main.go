package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/config"
	vpndata "github.com/privatix/dappctrl/svc/dappvpn/data"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/svc/dappvpn/msg"
	"github.com/privatix/dappctrl/svc/dappvpn/prepare"
	"github.com/privatix/dappctrl/tc"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/version"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

var (
	conf    *config.Config
	channel string
	logger  log.Logger
	tctrl   *tc.TrafficControl
	fatal   = make(chan string)
)

func createLogger() (log.Logger, io.Closer, error) {
	elog, err := log.NewStderrLogger(conf.FileLog.WriterConfig)
	if err != nil {
		return nil, nil, err
	}

	flog, closer, err := log.NewFileLogger(conf.FileLog)
	if err != nil {
		return nil, nil, err
	}

	logger := log.NewMultiLogger(elog, flog).Add("env", os.Environ())

	return logger, closer, nil
}

func main() {
	v := flag.Bool("version", false, "Prints current dappctrl version")

	fconfig := flag.String(
		"config", "dappvpn.config.json", "Configuration file")
	// Client mode is active when the parameter below is set.
	fchannel := flag.String("channel", "", "Channel ID for client mode")
	flag.Parse()

	version.Print(*v, Commit, Version)

	conf = config.NewConfig()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s\n", err))
	}

	channel = *fchannel

	var err error

	var closer io.Closer
	logger, closer, err = createLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}
	defer closer.Close()

	tctrl = tc.NewTrafficControl(conf.TC, logger)

	switch os.Getenv("script_type") {
	case "user-pass-verify":
		handleAuth()
	case "client-connect":
		handleConnect()
	case "client-disconnect":
		handleDisconnect()
	default:
		handleMonitor(*fconfig)
	}
}

func handleAuth() {
	user, pass := getCreds()
	args := sesssrv.AuthArgs{ClientID: user, Password: pass}

	err := sesssrv.Post(conf.Server.Config, logger, conf.Server.Username,
		conf.Server.Password, sesssrv.PathAuth, args, nil)
	if err != nil {
		logger.Fatal("failed to auth: " + err.Error())
	}

	if cn := commonNameOrEmpty(); len(cn) != 0 {
		storeChannel(cn, user)
	}
	storeChannel(user, user) // Needed when using username-as-common-name.
}

func handleConnect() {
	port, err := strconv.Atoi(os.Getenv("trusted_port"))
	if err != nil || port <= 0 || port > 0xFFFF {
		logger.Fatal("bad trusted_port value")
	}

	args := sesssrv.StartArgs{
		ClientID:   loadChannel(),
		ClientIP:   os.Getenv("trusted_ip"),
		ClientPort: uint16(port),
	}

	var res sesssrv.StartResult
	err = sesssrv.Post(conf.Server.Config, logger, conf.Server.Username,
		conf.Server.Password, sesssrv.PathStart, args, &res)
	if err != nil {
		logger.Fatal("failed to start session: " + err.Error())
	}

	if len(channel) != 0 || res.Offering.AdditionalParams == nil {
		return
	}

	var params vpndata.OfferingParams
	err = json.Unmarshal(res.Offering.AdditionalParams, &params)
	if err != nil {
		logger.Add("offering_params", res.Offering.AdditionalParams).Fatal(
			"failed to unmarshal offering params: " + err.Error())
	}

	err = tctrl.SetRateLimit(os.Getenv("dev"),
		os.Getenv("ifconfig_pool_remote_ip"),
		params.MinUploadMbits, params.MinDownloadMbits)
	if err != nil {
		logger.Fatal("failed to set rate limit: " + err.Error())
	}
}

func handleDisconnect() {
	down, err := strconv.ParseUint(os.Getenv("bytes_sent"), 10, 64)
	if err != nil || down < 0 {
		panic("bad bytes_sent value")
	}

	up, err := strconv.ParseUint(os.Getenv("bytes_received"), 10, 64)
	if err != nil || up < 0 {
		panic("bad bytes_received value")
	}

	args := sesssrv.StopArgs{
		ClientID: loadChannel(),
		Units:    down + up,
	}

	err = sesssrv.Post(conf.Server.Config, logger, conf.Server.Username,
		conf.Server.Password, sesssrv.PathStop, args, nil)
	if err != nil {
		logger.Fatal("failed to stop session: " + err.Error())
	}

	err = tctrl.UnsetRateLimit(os.Getenv("dev"),
		os.Getenv("ifconfig_pool_remote_ip"))
	if err != nil {
		logger.Fatal("failed to unset rate limit: " + err.Error())
	}
}

func handleMonStarted(ch string) bool {
	args := sesssrv.StartArgs{
		ClientID: ch,
	}

	err := sesssrv.Post(conf.Server.Config, logger, conf.Server.Username,
		conf.Server.Password, sesssrv.PathStart, args, nil)
	if err != nil {
		logger.Add("channel", ch).Error(
			"failed to start session for channel: " + err.Error())
		return false
	}

	return true
}

func handleMonStopped(ch string, up, down uint64) bool {
	args := sesssrv.StopArgs{
		ClientID: ch,
		Units:    down + up,
	}

	err := sesssrv.Post(conf.Server.Config, logger, conf.Server.Username,
		conf.Server.Password, sesssrv.PathStop, args, nil)
	if err != nil {
		logger.Add("channel", ch).Error(
			"failed to stop session for channel: " + err.Error())
		return false
	}

	return true
}

func handleMonByteCount(ch string, up, down uint64) bool {
	args := sesssrv.UpdateArgs{
		ClientID: ch,
		Units:    down + up,
	}

	err := sesssrv.Post(conf.Server.Config, logger, conf.Server.Username,
		conf.Server.Password, sesssrv.PathUpdate, args, nil)
	if err != nil {
		logger.Add("channel", ch).Error(
			"failed to update session for channel: " + err.Error())
		return false
	}

	return true
}

func handleMonitor(confFile string) {
	handleSession := func(ch string, event int, up, down uint64) bool {
		switch event {
		case mon.SessionStarted:
			return handleMonStarted(ch)
		case mon.SessionStopped:
			return handleMonStopped(ch, up, down)
		case mon.SessionByteCount:
			return handleMonByteCount(ch, up, down)
		default:
			return false
		}
	}

	dir := filepath.Dir(confFile)

	if len(channel) == 0 && !msg.IsDone(dir) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			pusher := msg.NewPusher(conf.Pusher,
				conf.Server.Config, conf.Server.Username,
				conf.Server.Password, logger)

			if err := pusher.PushConfiguration(ctx); err != nil {
				logger.Error("failed to push app config to" +
					" dappctrl: " + err.Error())
				return
			}

			if err := msg.Done(dir); err != nil {
				logger.Add("file", msg.PushedFile, "dir", dir,
					"err", err,
				).Error("failed to save file in directory")
			}
		}()
	}

	if len(channel) != 0 {
		if err := prepare.ClientConfig(logger, channel,
			conf); err != nil {
			logger.Fatal("failed to prepare client" +
				" configuration: " + err.Error())
		}

		ovpn := launchOpenVPN()
		defer ovpn.Kill()
		time.Sleep(conf.OpenVPN.StartDelay * time.Millisecond)
	}

	monitor := mon.NewMonitor(conf.Monitor, logger, handleSession, channel)
	go func() {
		fatal <- fmt.Sprintf("failed to monitor vpn traffic: %s",
			monitor.MonitorTraffic())
	}()

	logger.Fatal(<-fatal)
}

func launchOpenVPN() *os.Process {
	if len(conf.OpenVPN.Name) == 0 {
		logger.Fatal("no OpenVPN command provided")
	}

	args := append(conf.OpenVPN.Args, "--config")
	args = append(args, filepath.Join(
		conf.OpenVPN.ConfigRoot, channel, "client.ovpn"))

	cmd := exec.Command(conf.OpenVPN.Name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Fatal("failed to access OpenVPN stdout: " + err.Error())
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Fatal("failed to access OpenVPN stderr: " + err.Error())
	}

	if err := cmd.Start(); err != nil {
		logger.Fatal("failed to launch OpenVPN: " + err.Error())
	}

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			io.WriteString(
				os.Stderr, "openvpn: "+scanner.Text()+"\n")
		}
		if err := scanner.Err(); err != nil {
			msg := "failed to read from openVPN stdout/stderr: %s"
			fatal <- fmt.Sprintf(msg, err)
		}
		stdout.Close()
		stderr.Close()
	}()

	go func() {
		fatal <- fmt.Sprintf("OpenVPN exited: %v", cmd.Wait())
	}()

	return cmd.Process
}
