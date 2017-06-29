FROM golang:latest

RUN apt-get update && apt-get install -y jq

RUN mkdir -p /go/src/github.com/tendermint/basecoin
WORKDIR /go/src/github.com/tendermint/basecoin

COPY Makefile /go/src/github.com/tendermint/basecoin/
COPY glide.yaml /go/src/github.com/tendermint/basecoin/
COPY glide.lock /go/src/github.com/tendermint/basecoin/

RUN make get_vendor_deps

COPY . /go/src/github.com/tendermint/basecoin
