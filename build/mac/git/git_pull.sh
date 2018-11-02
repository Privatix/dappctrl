#!/usr/bin/env bash

. ${1}

for repository in ${REPOSITORIES[@]}
do
    echo "${repository}"
    cd "${repository}"
    git pull --all
done
