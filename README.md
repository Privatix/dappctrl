[![Go report](https://goreportcard.com/badge/github.com/Privatix/dappctrl)](https://goreportcard.com/badge/github.com/Privatix/dappctrl)
[![Maintainability](https://api.codeclimate.com/v1/badges/7e76f071e5408b13ea53/maintainability)](https://codeclimate.com/github/Privatix/dappctrl/maintainability)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FPrivatix%2Fdappctrl.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FPrivatix%2Fdappctrl?ref=badge_shield)

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
psql -U postgres -d postgres -c 'CREATE DATABASE dappctrl'
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
