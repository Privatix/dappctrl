#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl
DAPPCTRL_DIR=$HOME/go/src/${DAPPCTRL}

curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
cd "${DAPPCTRL_DIR}" && dep ensure
go get -d ${DAPPCTRL}/...
go get -u gopkg.in/reform.v1/reform
go get -u github.com/rakyll/statik
go get github.com/ethereum/go-ethereum/cmd/abigen

go generate ${DAPPCTRL}/...

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

export GIT_COMMIT
export GIT_RELEASE

go install -ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" -tags=notest ${DAPPCTRL}

