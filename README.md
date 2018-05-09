[![Go report](https://goreportcard.com/badge/github.com/Privatix/dappctrl)](https://goreportcard.com/badge/github.com/Privatix/dappctrl)
[![Maintainability](https://api.codeclimate.com/v1/badges/7e76f071e5408b13ea53/maintainability)](https://codeclimate.com/github/Privatix/dappctrl/maintainability)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FPrivatix%2Fdappctrl.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FPrivatix%2Fdappctrl?ref=badge_shield)

[develop](https://github.com/Privatix/dappctrl/tree/develop):
<img align="center" src="https://ci.privatix.net/plugins/servlet/wittified/build-status/PC-ICT0">

# Privatix Controller

Privatix Controller is a core of Agent and Client functionality.

# Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Prerequisites

Install prerequisite software if it's not installed.

* Install [Golang](https://golang.org/doc/install). Make sure that `$HOME/go/bin` (default `GOPATH`) is added to system path `$PATH`.

* Install [PostgreSQL](https://www.postgresql.org/download/).

* Install [gcc](https://gcc.gnu.org/install/).

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
go install github.com/ethereum/go-ethereum/cmd/abigen
go generate $DAPPCTRL/...
go install -tags=notest $DAPPCTRL
```

Prepare a `dappctrl` database instance:

```bash
psql -U postgres -f $DAPPCTRL_DIR/data/settings.sql
psql -U postgres -d dappctrl -f $DAPPCTRL_DIR/data/schema.sql
```

Make a copy of `dappctrl.config.json`:

```bash
cp dappctrl.config.json dappctrl.config.local.json
```

Modify `dappctrl.config.local.json` if you need non-default configuration and run:

```bash
dappctrl -config=$DAPPCTRL_DIR/dappctrl.config.local.json
```

Build `OpenVPN` service adapter:

```bash
go install $DAPPCTRL/svc/dappvpn
```

## Using docker

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

# Tests

## Preparing the test environment

1. Set variables for your test environment, e.g.:

    ```bash
    CONF_FILE=$DAPPCTRL_DIR/dappctrl-test.config.json
    LOCAL_CONF_FILE=$HOME/dappctrl-test.config.json
    DB_IP=10.16.194.21
    STRESS_JOBS=1000
    ```

2. Generate locally a configuration file using these variables, e.g.:

    ```bash
    jq ".DB.Conn.host=\"$DB_IP\" | .JobTest.StressJobs=$STRESS_JOBS" $CONF_FILE > $LOCAL_CONF_FILE
    ```

    **Note**: See `jq` [manual](https://stedolan.github.io/jq/manual) for
    syntax details.

## Running the tests

```bash
go test $DAPPCTRL/... -p=1 -config=$LOCAL_CONF_FILE
```

## Excluding specific tests from test run

It's possible to exclude arbitrary package tests from test runs. To do so use
a dedicated *build tag*. Name of a such tag is composed from the `no`-prefix,
name of the package and the `test` suffix. For example, using `noethtest` tag
will disable Ethereum library tests and disabling `novpnmontest` will disable
VPN monitor tests.

Example of a test run with the tags above:

```bash
go test $DAPPCTRL/... -p=1 -tags="noethtest nojobtest" -config=$LOCAL_CONF_FILE
```

# Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/Privatix/dappctrl/tags).

## Authors

* [ababo](https://github.com/ababo)
* [furkhat](https://github.com/furkhat)

See also the list of [contributors](https://github.com/Privatix/dappctrl/contributors) who participated in this project.

# License

This project is licensed under the **GPL-3.0 License** - see the [COPYING](COPYING) file for details.
