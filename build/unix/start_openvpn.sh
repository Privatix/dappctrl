#!/usr/bin/env bash

. ${1}

echo copy dappvpn config
cp ./bin/dapp_openvpn/dappvpn.client.config.json \
   ./bin/openvpn_client/config/dappvpn.config.json

cp ./bin/dapp_openvpn/dappvpn.agent.config.json \
   ./bin/openvpn_server/config/dappvpn.config.json

# CLIENT
echo install openvpn_client

cd ./bin/openvpn_client/bin/
sudo ./openvpn-inst install -config=../installer.config.json

# SERVER
echo install openvpn_server

cd ../../openvpn_server/bin/
sudo ./openvpn-inst install -config=../installer.config.json

echo start openvpn_server
sudo ./openvpn-inst start
