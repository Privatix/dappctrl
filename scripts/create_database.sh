#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/${DAPPCTRL}

psql -U postgres -f "${DAPPCTRL_DIR}/data/settings.sql"

dappctrl db-migrate -conn 'host=localhost sslmode=disable dbname=dappctrl user=postgres port=5433'
dappctrl db-init-data -conn 'host=localhost sslmode=disable dbname=dappctrl user=postgres port=5433'

