#!/usr/bin/env bash

MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPPCTRL_DIR="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dappctrl path

CONF_FILE="${DAPPCTRL_DIR}/dappctrl-test.config.json"
LOCAL_CONF_FILE=$HOME/dappctrl-test.config.json
DB_IP=10.16.194.21
STRESS_JOBS=1000

jq ".DB.Conn.host=\"$DB_IP\" | .JobTest.StressJobs=$STRESS_JOBS" "${CONF_FILE}" > "${LOCAL_CONF_FILE}"
go test ${DAPPCTRL}/... -p=1 -config="${LOCAL_CONF_FILE}"
