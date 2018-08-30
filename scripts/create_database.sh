#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/${DAPPCTRL}

psql -U postgres -f "${DAPPCTRL_DIR}/data/settings.sql"

echo
echo "Database has been created successfully."
