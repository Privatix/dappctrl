We have prepared two images and a compose file to make it easier to run app and its dependencies.

There are 3 services in compose file:

1. `db` — uses public `postgres` image
1. `vpn` — image `privatix/dapp-vpn-server` is an openvpn with attached
`dappvpn`.
1. `dappctrl` — image `privatix/dappctrl` is a main controller app

If you want to develop `dappctrl` then it is convenient to run its dependencies using `docker`, but controller itself at your host machine:

```
docker-compose up vpn db
```

If your app is using `dappctrl` or you are not planning to develop controller run

```
docker-compose up
```
