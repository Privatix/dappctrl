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
RUN echo "port 1194" >> server.conf
RUN echo "proto udp" >> server.conf
RUN echo "dev tun" >> server.conf
RUN echo "ca ca.crt" >> server.conf
RUN echo "cert myvpn.crt" >> server.conf
RUN echo "key myvpn.key" >> server.conf
RUN echo "dh dh2048.pem" >> server.conf
RUN echo "server 10.0.0.0 255.255.255.0" >> server.conf
RUN echo "ifconfig-pool-persist ipp.txt" >> server.conf
RUN echo "keepalive 10 120" >> server.conf
RUN echo "tls-auth ta.key 0" >> server.conf
RUN echo "cipher AES-256-CBC" >> server.conf
RUN echo "persist-key" >> server.conf
RUN echo "persist-tun" >> server.conf
RUN echo "status /var/log/openvpn-status.log" >> server.conf
RUN echo "verb 3" >> server.conf
RUN echo "explicit-exit-notify 1" >> server.conf
## allow management console used by dappctrl monitor
RUN echo "management 0.0.0.0 7505" >> server.conf
## link to dapptrig
RUN echo "auth-user-pass-verify /go/bin/dapptrig via-file" >> server.conf
RUN echo "client-connect /go/bin/dapptrig" >> server.conf
RUN echo "client-disconnect /go/bin/dapptrig" >> server.conf
RUN echo "script-security 3" >> server.conf

# expose ports
EXPOSE 7505
EXPOSE 1194

# run at image start
WORKDIR /etc/openvpn/config
CMD [ "openvpn", "server.conf" ]
