FROM golang:1.25.5-alpine3.23 AS buildbase

RUN apk add build-base

WORKDIR /go/src/github.com/distraw/transaction-indexer

COPY . .

ENV CGO_ENABLED=1
ENV GOOS="linux"

RUN go build -o /usr/local/bin/transaction-indexer /go/src/github.com/distraw/transaction-indexer

FROM alpine:3.23

COPY --from=buildbase /usr/local/bin/transaction-indexer /usr/local/bin/transaction-indexer
COPY --from=buildbase /go/src/github.com/distraw/transaction-indexer/config.local.yaml /usr/local/bin/config.yaml

ENTRYPOINT [ "transaction-indexer" ]