# Branch/tag, latest active release by default, set to specific release
ARG STELLAR_CORE_VERSION="v11.3.0"

# ===============================================================================================
FROM astroband/docker-stellar-core-alpine:${STELLAR_CORE_VERSION} AS stellar-core

# ===============================================================================================

FROM golang:alpine AS build

RUN apk add --no-cache git build-base

RUN mkdir -p $GOPATH/src/github.com/astroband/stellar-parallel-catchup
WORKDIR $GOPATH/src/github.com/astroband/stellar-parallel-catchup

ADD . .

RUN GO111MODULE=on go build

# ===============================================================================================

FROM alpine:latest

ENV DATABASE_URL=postgres://localhost/core?sslmode=disable
ENV MAX_LEDGER=99999999
ENV CONCURRENCY=3

WORKDIR /root

RUN apk add --no-cache libpq libstdc++

COPY --from=stellar-core /usr/local/bin/stellar-core .
COPY --from=build /go/src/github.com/astroband/stellar-parallel-catchup/stellar-parallel-catchup .
COPY --from=build /go/src/github.com/astroband/stellar-parallel-catchup/templates/stellar-core.cfg ./stellar-core.cfg
RUN chmod +x ./stellar-parallel-catchup

CMD ["/root/stellar-parallel-catchup"]