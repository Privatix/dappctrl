package ovpn

const ClientConfig = `# tunX | tapX | null TUN/TAP virtual network device
( X can be omitted for a dynamic device.)
dev tun

# Use protocol tcp for communicating
# with remote host
proto {{.Proto}} 

# Encrypt packets with AES-256-CBC algorithm
cipher {{.Cipher}}

# Enable TLS and assume client role
# during TLS handshake.
tls-client

client

# Remote host name or IP address
# with port number and protocol tcp
# for communicating
remote {{.ServerAddress}} {{.Port}} {{.Proto}} 

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
ping-restart {{.PingRestart}}

# Ping remote over the TCP/UDP control
# channel if no packets have been sent for
# at least n seconds
ping {{.Ping}} 


# The keepalive directive causes ping-like
# messages to be sent back and forth over
# the link so that each side knows when
# the other side has gone down.
# Ping every 10 seconds, assume that remote
# peer is down if no ping received during
# a 120 second time period.
keepalive 10 120

# Authenticate with server using
# username/password in interactive mode
auth-user-pass

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
connect-retry {{.ConnectRetry}} 

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
{{.CompLZO}} 

# Set log file verbosity.
verb 3
log-append /var/log/openvpn/openvpn-tcp.log
`
