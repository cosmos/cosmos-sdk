FROM golang:1.15-alpine as build

# required due to rosetta-cli's ethereum's dependencies
RUN apk add git gcc libc-dev --no-cache

ARG ROSETTA_VERSION="v0.6.0"

# build simd
WORKDIR /simd
COPY . ./
RUN go build -o simd ./simapp/simd/

# build rosetta CLI
WORKDIR /rosetta
RUN git clone https://github.com/coinbase/rosetta-cli
WORKDIR /rosetta/rosetta-cli
RUN git checkout tags/$ROSETTA_VERSION
RUN go build -o rosetta-cli ./main.go

FROM alpine

RUN apk add gcc libc-dev --no-cache

ENV PATH=$PATH:/bin

WORKDIR /bin
COPY --from=build /simd/simd /bin/simd
COPY --from=build /rosetta/rosetta-cli/rosetta-cli /bin/rosetta-cli

WORKDIR /rosetta
COPY ./contrib/rosetta/configuration ./

WORKDIR /root/.simapp

COPY ./contrib/rosetta/node/data ./

