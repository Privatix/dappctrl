FROM privatix/dappctrl

# prepare and build dapptrig

ARG APP=github.com/privatix/dappctrl/tool/dapptrig
ARG APP_HOME=/go/src/${APP}
WORKDIR $APP_HOME

## build
RUN go install -tags=notest ${APP}

# install openvpn

RUN apk --no-cache add \
	openvpn \
    easy-rsa

## create cert files
WORKDIR /usr/share/easy-rsa
RUN cp vars.example vars
RUN ./easyrsa init-pki
RUN echo "my-common-name" | ./easyrsa build-ca nopass
RUN ./easyrsa build-server-full myvpn nopass
RUN openvpn --genkey --secret ta.key

## copy cert files to config directory
RUN mkdir /etc/openvpn/config
RUN cp ./pki/ca.crt /etc/openvpn/config/
RUN cp ./pki/issued/myvpn.crt /etc/openvpn/config/
RUN cp ./pki/private/myvpn.key /etc/openvpn/config/


WORKDIR /etc/openvpn/config
RUN openssl dhparam -out dh2048.pem 2048
RUN openvpn --genkey --secret ta.key

## create config
RUN echo $'                                           \n\
port 1194                                             \n\
proto udp                                             \n\
dev tun                                               \n\
ca ca.crt                                             \n\
cert myvpn.crt                                        \n\
key myvpn.key                                         \n\
dh dh2048.pem                                         \n\
server 10.0.0.0 255.255.255.0                         \n\
ifconfig-pool-persist ipp.txt                         \n\
keepalive 10 120                                      \n\
tls-auth ta.key 0                                     \n\
cipher AES-256-CBC                                    \n\
persist-key                                           \n\
persist-tun                                           \n\
status /var/log/openvpn-status.log                    \n\
verb 3                                                \n\
explicit-exit-notify 1                                \n\
## allow management console used by dappctrl monitor  \n\
management 0.0.0.0 7505                               \n\
## link to dapptrig                                   \n\
auth-user-pass-verify /go/bin/dapptrig via-file       \n\
client-connect /go/bin/dapptrig                       \n\
client-disconnect /go/bin/dapptrig                    \n\
script-security 3                                     \n\
' >> server.conf

# expose ports
EXPOSE 7505
EXPOSE 1194

# run at image start
WORKDIR /etc/openvpn/config
CMD [ "openvpn", "server.conf" ]
