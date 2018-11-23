#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

port=${POSTGRES_PORT:-5432}
user=${POSTGRES_USER:-postgres}

connection_string="host=localhost sslmode=disable \
user=${user} \
port=${port} \
${POSTGRES_PASSWORD:+ password=${POSTGRES_PASSWORD}}"

echo Connection string: "${connection_string}"

dappctrl db-create -conn "${connection_string} dbname=postgres"
dappctrl db-migrate -conn "${connection_string} dbname=dappctrl"
dappctrl db-init-data -conn "${connection_string} dbname=dappctrl"
