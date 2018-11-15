#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

if [ -z "${POSTGRES_PORT}" ]
then
    POSTGRES_PORT=5432
fi
if [ -z "${DAPPCTRL_DIR}" ]
then
    DAPPCTRL_DIR=${GOPATH}/src/${DAPPCTRL}
fi

psql -U postgres -f "${DAPPCTRL_DIR}/data/settings.sql"

dappctrl db-migrate -conn 'host=localhost sslmode=disable dbname=dappctrl user=postgres port='${POSTGRES_PORT}
dappctrl db-init-data -conn 'host=localhost sslmode=disable dbname=dappctrl user=postgres port='${POSTGRES_PORT}

