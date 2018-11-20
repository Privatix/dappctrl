#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

if [ -z "${DAPPCTRL_DIR}" ]
then
    DAPPCTRL_DIR=${GOPATH}/src/${DAPPCTRL}
fi

if [ ! -f "${GOPATH}"/bin/dep ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
fi
echo running dep ensure
cd "${DAPPCTRL_DIR}" && dep ensure
go get -d ${DAPPCTRL}/...
go get -u gopkg.in/reform.v1/reform
go get -u github.com/rakyll/statik
go get -u github.com/pressly/goose/cmd/goose
go get github.com/ethereum/go-ethereum/cmd/abigen

go generate ${DAPPCTRL}/...

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

export GIT_COMMIT
export GIT_RELEASE

go install -ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" -tags=notest ${DAPPCTRL}
