# tunX | tapX | null TUN/TAP virtual network device
# ( X can be omitted for a dynamic device.)
dev tun

# Use protocol tcp for communicating
# with remote host
proto {{if .Proto}}{{.Proto}}{{else}}tcp{{autogen}}{{end}}

# Encrypt packets with AES-256-CBC algorithm
cipher {{if .Cipher}}{{.Cipher}}{{else}}AES-256-CBC{{autogen}}{{end}}

# Enable TLS and assume client role
# during TLS handshake.
tls-client

client

# Remote host name or IP address
# with port number and protocol tcp
# for communicating
{{if .ServerAddress}}{{if .Port}}remote {{.ServerAddress}} {{.Port}} {{.Proto}}{{end}}{{end}}

# If hostname resolve fails for --remote,
# retry resolve for n seconds before failing.
# Set n to "infinite" to retry indefinitely.
resolv-retry infinite

# Do not bind to local address and port.
# The IP stack will allocate a dynamic
# port for returning packets.
# Since the value of the dynamic port
# could not be known in advance by a peer,
# this option is only suitable for peers
# which will be initiating connections
# by using the --remote option.
nobind

# Don't close and reopen TUN/TAP device
# or run up/down scripts across SIGUSR1
# or --ping-restart restarts.
# SIGUSR1 is a restart signal similar
# to SIGHUP, but which offers
# finer-grained control over reset options.
persist-tun

# Don't re-read key files across
# SIGUSR1 or --ping-restart.
persist-key

# Trigger a SIGUSR1 restart after n seconds
# pass without reception of a ping
# or other packet from remote.
ping-restart {{if .PingRestart}}{{.PingRestart}}{{else}}10{{autogen}}{{end}}

# Ping remote over the TCP/UDP control
# channel if no packets have been sent for
# at least n seconds
ping {{if .Ping}}{{.Ping}}{{else}}10{{autogen}}{{end}}

# Authenticate with server using
# username/password in interactive mode
auth-user-pass access.ovpn

# or you can use directive below:
# auth-user-pass /path/to/pass.file Authenticate
# with server using username/password.
# /path/to/pass.file is a file
# containing username/password on 2 lines
#(Note: OpenVPN will only read passwords
# from a file if it has been
# built with the --enable-password-save
# configure option)

# Client will retry the connection
# without requerying for an
# --auth-user-pass username/password.
auth-retry nointeract

# Become a daemon after all initialization
# functions are completed. This option will
# cause all message and error output
# to be sent to the log file
daemon

# take n as the number of seconds
# to wait between connection retries
connect-retry {{if .ConnectRetry}}{{.ConnectRetry}}{{else}}2 120{{autogen}}{{end}}

# uncomment this section
# if you want use ca.crt file
;ca /path/to/ca.crt
# or you can include ca certificate
# in this file like a below:
<ca>
{{.Ca}}</ca>

# Enable compression on the VPN link.
# Don't enable this unless it is also
# enabled in the server config file.
# Use fast LZO compression -- may add up
# to 1 byte per packet for incompressible data.
{{if .CompLZO}}{{.CompLZO}}{{else}};comp-lzo{{end}}

# Set log file verbosity.
verb 3
log-append /var/log/openvpn/openvpn-tcp.log
