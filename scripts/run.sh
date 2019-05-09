#!/usr/bin/env bash

MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPPCTRL_DIR="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dappctrl path

dappctrl -config="${DAPPCTRL_DIR}/dappctrl.config.local.json"
