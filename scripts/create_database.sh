#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

port=${POSTGRES_PORT:-5432}
user=${POSTGRES_USER:-postgres}

connection_string="host=localhost sslmode=disable user=${user} port=${port}"

if [[ "${POSTGRES_PASSWORD}" ]]
then
   connection_string="${connection_string} password=${POSTGRES_PASSWORD}"
fi

echo Connection string: "${connection_string}"

dappctrl db-create -conn "${connection_string} dbname=postgres"
dappctrl db-migrate -conn "${connection_string} dbname=dappctrl"
dappctrl db-init-data -conn "${connection_string} dbname=dappctrl"
