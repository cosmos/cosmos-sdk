FROM golang:1.15-alpine as build

ARG COSMOS_VERSION="v0.40.0-rc2"

RUN apk add git --no-cache

WORKDIR /build
# build
RUN git clone https://github.com/cosmos/cosmos-sdk
WORKDIR /build/cosmos-sdk
RUN git checkout $COSMOS_VERSION
RUN CGO_ENABLED=0 go build -o simd ./simapp/simd/

FROM alpine

ENV PATH=$PATH:/bin

WORKDIR /bin
COPY --from=build /build/cosmos-sdk/simd ./simd

WORKDIR /root/.simapp

COPY ./server/rosetta/test/data/node ./

