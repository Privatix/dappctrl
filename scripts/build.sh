#!/usr/bin/env bash

MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPPCTRL="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dappctrl path

echo ${DAPPCTRL}

echo
echo go get
echo

cd "${DAPPCTRL}"

go get -u -v gopkg.in/reform.v1/reform
go get -u -v github.com/rakyll/statik
go get -u -v github.com/pressly/goose/cmd/goose
go get -v github.com/ethereum/go-ethereum/cmd/abigen

echo
echo go generate
echo

go generate -x ${DAPPCTRL}/...

GIT_COMMIT=$(git rev-list -1 HEAD | head -n 1)

if [ -z ${VERSION_TO_SET_IN_BUILDER} ]; then
    GIT_RELEASE=$(git tag -l --points-at HEAD | head -n 1)
    # if $GIT_RELEASE is zero:
    GIT_RELEASE=${GIT_RELEASE:-$(git rev-parse --abbrev-ref HEAD | grep -o "[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}")}
else
    GIT_RELEASE=${VERSION_TO_SET_IN_BUILDER}
fi

echo
echo GIT_COMMIT=${GIT_COMMIT}
echo GIT_RELEASE=${GIT_RELEASE}

export GIT_COMMIT
export GIT_RELEASE

echo
echo go install
echo

if [[ ! -d "${GOPATH}/bin/" ]]; then
    mkdir "${GOPATH}/bin/" || exit 1
fi

echo $GOPATH/bin/dappctrl
go install -ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" \
    -tags=notest ${DAPPCTRL} || exit 1


echo
echo done
