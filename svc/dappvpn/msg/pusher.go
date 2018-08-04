package msg

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/rdegges/go-ipify"

	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

const (
	defaultIP              = "127.0.0.1"
	serverAddressParameter = "externalIP"
	caDataParameter        = "caData"

	// PushedFile the name of a file that indicates that
	// the configuration is already loaded on the server.
	PushedFile = "configPushed"
	filePerm   = 0644
)

// Config is configuration to Pusher.
type Config struct {
	ExportConfigKeys []string
	ConfigPath       string
	CaCertPath       string
	TimeOut          int64
}

// Pusher updates the product configuration.
type Pusher struct {
	config   *Config
	server   *srv.Config
	username string
	password string
	logger   log.Logger
	ip       string
}

// NewConfig creates a default configuration.
func NewConfig() *Config {
	return &Config{}
}

// NewPusher creates a new Pusher object.
// Argument conf to parsing vpn configuration. Arguments srv, user, pass
// to send configuration to session service.
func NewPusher(conf *Config, srv *srv.Config, user, pass string,
	logger log.Logger) *Pusher {
	var ip string
	ip, err := externalIP()
	if err != nil {
		logger.Warn("couldn't get my IP address")
		ip = defaultIP
	}

	return &Pusher{
		config:   conf,
		server:   srv,
		username: user,
		password: pass,
		logger:   logger,
		ip:       ip,
	}
}

func (p *Pusher) vpnParams() (map[string]string, error) {
	vpnParams, err := vpnParams(p.logger, p.config.ConfigPath,
		p.config.ExportConfigKeys)
	if err != nil {
		return nil, err
	}

	ca, err := certificateAuthority(p.logger, p.config.CaCertPath)
	if err != nil {
		return nil, err
	}

	vpnParams[serverAddressParameter] = p.ip
	vpnParams[caDataParameter] = string(ca)

	return vpnParams, err
}

// PushConfiguration send the vpn configuration to session server.
func (p *Pusher) PushConfiguration(ctx context.Context) error {
	params, err := p.vpnParams()
	if err != nil {
		return err
	}

	args := &sesssrv.ProductArgs{
		Config: params,
	}

	for {
		select {
		case <-ctx.Done():
			return ErrContextIsDone
		default:
		}

		if err := sesssrv.Post(p.server, p.username, p.password,
			sesssrv.PathProductConfig,
			*args, nil); err != nil {
			p.logger.Add("error", err.Error()).Warn(
				"failed to push app config to dappctrl.")
			time.Sleep(time.Second *
				time.Duration(p.config.TimeOut))
			continue
		}
		p.logger.Info("vpn server configuration has been" +
			" successfully sent to dappctrl")
		break
	}
	return nil
}

func externalIP() (string, error) {
	return ipify.GetIp()
}

// IsDone checks if the vpn configuration is loaded to server.
func IsDone(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, PushedFile))
	return err == nil
}

// Done makes configPushed file.
func Done(dir string) error {
	file := filepath.Join(dir, PushedFile)
	return ioutil.WriteFile(file, nil, filePerm)
}
