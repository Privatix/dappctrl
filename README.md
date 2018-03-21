[![Go report](https://goreportcard.com/badge/github.com/Privatix/dappctrl)](https://goreportcard.com/badge/github.com/Privatix/dappctrl)
[![Maintainability](https://api.codeclimate.com/v1/badges/7e76f071e5408b13ea53/maintainability)](https://codeclimate.com/github/Privatix/dappctrl/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/7e76f071e5408b13ea53/test_coverage)](https://codeclimate.com/github/Privatix/dappctrl/test_coverage)

# Privatix Controller.
Privatix Controller is a core of Agent and Client functionality.

# Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Prerequisites

Install prerequisite software:
* Install Golang if it's not installed. Make sure that `$HOME/go/bin` is added
to `$PATH`.
* Install [PostgreSQL](https://www.postgresql.org/download/) if it's not installed.

## Installation steps

Clone the `dappctrl` repository using git:

```
git clone https://github.com/Privatix/dappctrl.git
cd dappctrl
git checkout master
```

Build `dappctrl` package:

```bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/$DAPPCTRL
mkdir -p $DAPPCTRL_DIR
git clone git@github.com:Privatix/dappctrl.git $DAPPCTRL_DIR
go get -d $DAPPCTRL/...
go get -u gopkg.in/reform.v1/reform
go generate $DAPPCTRL
go install -tags=notest $DAPPCTRL
```

Prepare a `dappctrl` database instance:

```bash
psql -U postgres -c 'CREATE DATABASE dappctrl'
psql -U postgres -d dappctrl -f $DAPPCTRL_DIR/data/schema.sql
psql -U postgres -d dappctrl -f $DAPPCTRL_DIR/data/test_data.sql
```

Modify `dappctrl.config.json` if you need non-default configuration and run:

```bash
dappctrl -config=$DAPPCTRL_DIR/dappctrl.config.json
```

Build OpenVPN session trigger:

```bash
go install $DAPPCTRL/tool/dapptrig
```

## Using docker
We have prepared two images and a compose file to make it easier to run app and its dependencies.
There are 3 services in compose file:
  1. db — uses public postgres image
  2. vpn — image `privatix/dapp-vpn-server` is an openvpn that uses `dapptrig` for auth, connect and disconnect
  3. dappctrl — image `privatix/dappctrl` is a main controller app

If you want to develop dappctrl then it is convenient to run its dependencies using docker, but controller itself at your host machine:
```
docker-compose up vpn db
```
If your app is using dappctrl or you are not planning to develop controller run
```
docker-compose up
```

# Tests

## Running the tests

```bash
go test $DAPPCTRL/... -config=$DAPPCTRL_DIR/dappctrl-test.config.json
```

# Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/Privatix/dapp-somc/tags).

## Authors

* [ababo](https://github.com/ababo)
* [HaySayCheese](https://github.com/HaySayCheese)
* [furkhat](https://github.com/furkhat)

See also the list of [contributors](https://github.com/Privatix/dapp-somc/contributors) who participated in this project.

# License

This project is licensed under the **GPL-3.0 License** - see the [COPYING](COPYING) file for details.
