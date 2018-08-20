#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/${DAPPCTRL}

CONF_FILE="${DAPPCTRL_DIR}/dappctrl-test.config.json"
LOCAL_CONF_FILE=$HOME/dappctrl-test.config.json
DB_IP=10.16.194.21
STRESS_JOBS=1000

jq ".DB.Conn.host=\"$DB_IP\" | .JobTest.StressJobs=$STRESS_JOBS" "${CONF_FILE}" > "${LOCAL_CONF_FILE}"
go test ${DAPPCTRL}/... -p=1 -config="${LOCAL_CONF_FILE}"
