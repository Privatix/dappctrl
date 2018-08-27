#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/${DAPPCTRL}

dappctrl -config="${DAPPCTRL_DIR}/dappctrl.config.local.json"
