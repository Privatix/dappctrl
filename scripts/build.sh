#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

echo ${DAPPCTRL_DIR:=${GOPATH}/src/${DAPPCTRL}}

if [ ! -f "${GOPATH}"/bin/dep ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
fi

echo
echo dep ensure
echo

cd "${DAPPCTRL_DIR}"
rm -f Gopkg.lock
dep ensure -v

cd "${DAPPCTRL_DIR}/vendor/gopkg.in/reform.v1/reform" && go install .
cd "${DAPPCTRL_DIR}/vendor/github.com/rakyll/statik" && go install .
cd "${DAPPCTRL_DIR}/vendor/github.com/pressly/goose/cmd/goose" && go install .
cd "${DAPPCTRL_DIR}/vendor/github.com/ethereum/go-ethereum/cmd/abigen" && go install .
cd "${DAPPCTRL_DIR}"

echo
echo go generate
echo

go generate -x ${DAPPCTRL}/...

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

export GIT_COMMIT
export GIT_RELEASE

echo
echo go install
echo

echo $GOPATH/bin/dappctrl
go install -ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" \
    -tags=notest ${DAPPCTRL} || exit 1

echo
echo done
