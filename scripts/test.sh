#!/usr/bin/env bash

MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPPCTRL_DIR="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dappctrl path

go test "${DAPPCTRL_DIR}/..." -config="${DAPPCTRL_DIR}/dappctrl-test.config.json" -tags=noethtest -vv -p=1
