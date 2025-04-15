FROM golang:1.21-alpine3.17 AS build-env

RUN apk add --no-cache git gcc musl-dev && \
    apk add --update make && \
    go install github.com/google/wire/cmd/wire@latest

WORKDIR /go/src/github.com/devtron-labs/git-sensor

ADD . /go/src/github.com/devtron-labs/git-sensor/

RUN GOOS=linux make

FROM alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099

COPY ./git-ask-pass.sh /git-ask-pass.sh

RUN chmod +x /git-ask-pass.sh && \
	apk add --no-cache ca-certificates && \
        apk add tini --no-cache && \
	apk add git --no-cache && \
	apk add openssh --no-cache && \
	adduser -D devtron

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/git-sensor/scripts/ /go/src/github.com/devtron-labs/git-sensor/git-sensor ./
USER devtron
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./git-sensor"]
