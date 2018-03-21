FROM golang:alpine

ARG APP=github.com/privatix/dappctrl
ARG APP_HOME=/go/src/${APP}

RUN mkdir -p ${APP_HOME}

WORKDIR ${APP_HOME}

# copy app files
COPY . .

# build
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev
RUN go get -u gopkg.in/reform.v1/reform
RUN go get -d ${APP}/...
RUN go generate ${APP}
RUN go install -tags=notest ${APP}

CMD [ "dappctrl" ]
