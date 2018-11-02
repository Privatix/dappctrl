#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/${DAPPCTRL}

if [ -z "${POSTGRES_PORT}" ]
then
    POSTGRES_PORT=5432
fi

echo psql -U postgres -f "${DAPPCTRL_DIR}/data/settings.sql"

echo dappctrl db-migrate -conn 'host=localhost sslmode=disable dbname=dappctrl user=postgres port="'${POSTGRES_PORT}'"'
echo dappctrl db-init-data -conn 'host=localhost sslmode=disable dbname=dappctrl user=postgres port="'${POSTGRES_PORT}'"'

