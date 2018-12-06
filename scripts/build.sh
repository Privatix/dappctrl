#!/usr/bin/env bash
DAPPCTRL=github.com/privatix/dappctrl

echo ${DAPPCTRL_DIR:=${GOPATH}/src/${DAPPCTRL}}

if [ ! -f "${GOPATH}"/bin/dep ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
fi

echo
echo dep ensure
echo

cd "${DAPPCTRL_DIR}" && dep ensure -v

echo
echo go get
echo

go get -d -v ${DAPPCTRL}/...
go get -u -v gopkg.in/reform.v1/reform
go get -u -v github.com/rakyll/statik
go get -u -v github.com/pressly/goose/cmd/goose
go get -v github.com/ethereum/go-ethereum/cmd/abigen

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
