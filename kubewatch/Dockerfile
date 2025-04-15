FROM golang:1.21-alpine3.19  AS build-env

RUN echo $GOPATH && \
    apk add --no-cache git gcc musl-dev && \
    apk add --update make

WORKDIR /go/src/github.com/devtron-labs/kubewatch
ADD . /go/src/github.com/devtron-labs/kubewatch
RUN GOOS=linux make

FROM alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099

RUN apk add --update ca-certificates && \
    adduser -D devtron

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/kubewatch/kubewatch .

RUN chmod +x ./kubewatch

USER devtron

ENTRYPOINT ["./kubewatch"]
