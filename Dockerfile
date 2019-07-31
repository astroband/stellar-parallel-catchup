# ===============================================================================================

FROM alpine:latest AS stellar-core

# Branch/tag, latest by default
ARG STELLAR_CORE_VERSION="prod"

# https://gist.github.com/Globegitter/d75a39f7f7c0f6305ce52f3196c46cdc
RUN apk add --no-cache --virtual=pandoc-deps wget ca-certificates && \
  wget https://gitlab.com/ConorIA/alpine-pandoc/raw/master/conor@conr.ca-584aeee5.rsa.pub -O /etc/apk/keys/conor@conr.ca-584aeee5.rsa.pub && \
  echo https://conoria.gitlab.io/alpine-pandoc/ >> /etc/apk/repositories && \
  apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing cmark && \
  apk add --no-cache openssh git build-base pkgconfig \
  autoconf automake libtool bison flex postgresql-dev pandoc linux-headers && \
  apk del pandoc-deps

RUN git clone --branch ${STELLAR_CORE_VERSION} --recursive --depth 1 https://github.com/stellar/stellar-core.git

RUN cd stellar-core && ./autogen.sh && ./configure && make && make install

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

COPY --from=stellar-core /usr/local/bin/stellar-core .
COPY --from=build /go/src/github.com/astroband/stellar-parallel-catchup/stellar-parallel-catchup .
COPY --from=build /go/src/github.com/astroband/stellar-parallel-catchup/templates/stellar-core.cfg ./stellar-core.cfg
RUN chmod +x ./stellar-parallel-catchup

CMD ["/root/stellar-parallel-catchup"]