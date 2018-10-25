[![Go report](https://goreportcard.com/badge/github.com/Privatix/dappctrl)](https://goreportcard.com/badge/github.com/Privatix/dappctrl)
[![Maintainability](https://api.codeclimate.com/v1/badges/7e76f071e5408b13ea53/maintainability)](https://codeclimate.com/github/Privatix/dappctrl/maintainability)
[![GoDoc](https://godoc.org/github.com/Privatix/dappctrl?status.svg)](https://godoc.org/github.com/Privatix/dappctrl)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FPrivatix%2Fdappctrl.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FPrivatix%2Fdappctrl?ref=badge_shield)

[develop](https://github.com/Privatix/dappctrl/tree/develop):
<img align="center" src="https://ci.privatix.net/plugins/servlet/wittified/build-status/PC-ICT0">

# Privatix Controller

Privatix Controller is a core of Agent and Client functionality.

# Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Prerequisites

Install prerequisite software if it's not installed.

* Install [Golang](https://golang.org/doc/install). Make sure that `$GOPATH/bin` is added to system path `$PATH`.

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
/scripts/build.sh
```

Prepare a `dappctrl` database instance:

```bash
/scripts/create_database.sh
```

Make a copy of `dappctrl.config.json`:

```bash
cp dappctrl.config.json dappctrl.config.local.json
```

Modify `dappctrl.config.local.json` if you need non-default configuration and run:

```bash
/scripts/run.sh
```

For developing purposes, you have to use `dappctrl-dev.config.json`.

More information about `dappctrl.config.json`: [config fields description](https://github.com/Privatix/dappctrl/wiki/dappctrl.config.json-description).

## Building and configuring service adapters

* **OpenVPN** - please read `svc/dappvpn/README.md`.

## Using docker

You can use docker to simplify work with `dappctrl`: [using docker](https://github.com/Privatix/dappctrl/wiki/Using-docker)

# Tests

To run the tests execute following script:
```bash
/scripts/test.sh
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
* [dzeckelev](https://github.com/dzeckelev)

See also the list of [contributors](https://github.com/Privatix/dappctrl/contributors) who participated in this project.

# License

This project is licensed under the **GPL-3.0 License** - see the [COPYING](COPYING) file for details.
