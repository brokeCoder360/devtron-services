FROM golang:1.21-alpine3.18 AS build-env

RUN apk add --no-cache git gcc musl-dev && \
	apk add --update make

WORKDIR /go/src/github.com/devtron-labs/lens

ADD . /go/src/github.com/devtron-labs/lens

RUN go install github.com/google/wire/cmd/wire@latest && \
	GOOS=linux make

FROM alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099

RUN apk add --no-cache ca-certificates && \
	adduser -D devtron

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/lens/lens /go/src/github.com/devtron-labs/lens/scripts/ .

USER devtron

CMD ["./lens"]
