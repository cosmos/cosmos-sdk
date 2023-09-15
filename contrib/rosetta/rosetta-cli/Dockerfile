FROM golang:1.19-alpine as build

RUN apk add git gcc libc-dev --no-cache

ARG ROSETTA_VERSION="v0.6.7"

# build rosetta CLI
WORKDIR /rosetta
RUN git clone https://github.com/coinbase/rosetta-cli .
RUN git checkout tags/$ROSETTA_VERSION
RUN go build -o rosetta-cli ./main.go

FROM alpine
RUN apk add gcc libc-dev python3 --no-cache

ENV PATH=$PATH:/bin

COPY --from=build /rosetta/rosetta-cli /bin/rosetta-cli
