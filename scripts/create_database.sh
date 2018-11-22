#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

if [ -z "${POSTGRES_PORT}" ]
then
    POSTGRES_PORT=5432
fi

if [ -z "${POSTGRES_USER}" ]
then
    POSTGRES_USER=postgres
fi

if [ -z "${POSTGRES_PASSWORD}" ]
then
    POSTGRES_PASSWORD=
fi

if [ -z "${DAPPCTRL_DIR}" ]
then
    DAPPCTRL_DIR=${GOPATH}/src/${DAPPCTRL}
fi


dappctrl db-create -conn \
    'host=localhost sslmode=disable dbname=postgres user='${POSTGRES_USER}' port='${POSTGRES_PORT}' password='${POSTGRES_PASSWORD}

dappctrl db-migrate -conn \
    'host=localhost sslmode=disable dbname=dappctrl user='${POSTGRES_USER}' port='${POSTGRES_PORT}' password='${POSTGRES_PASSWORD}

dappctrl db-init-data -conn \
    'host=localhost sslmode=disable dbname=dappctrl user='${POSTGRES_USER}' port='${POSTGRES_PORT}' password='${POSTGRES_PASSWORD}

