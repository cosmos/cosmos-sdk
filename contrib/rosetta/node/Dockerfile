FROM golang:1.15-alpine as build

# required due to rosetta-cli's ethereum's dependencies
RUN apk add git gcc libc-dev --no-cache

RUN apk add --no-cache tar

ARG ROSETTA_VERSION="v0.5.23"

# prepare node data
WORKDIR /node
COPY ./contrib/rosetta/node/data.tar.gz data.tar.gz
RUN tar -zxvf data.tar.gz -C .

# build simd and simcli
WORKDIR /simd
COPY . ./
RUN go build -o simd ./contrib/rosetta/simd/cmd/simd/simd.go
RUN go build -o simcli ./contrib/rosetta/simd/cmd/simcli/simcli.go


# build rosetta CLI
WORKDIR /rosetta
RUN git clone https://github.com/coinbase/rosetta-cli .
RUN git checkout tags/$ROSETTA_VERSION
RUN go build -o rosetta-cli ./main.go

FROM alpine
RUN apk add gcc libc-dev python3 --no-cache

ENV PATH=$PATH:/bin

COPY --from=build /simd/simd /bin/simd
COPY --from=build /simd/simcli /bin/simcli
COPY --from=build /rosetta/rosetta-cli /bin/rosetta-cli

WORKDIR /rosetta
COPY ./contrib/rosetta/configuration ./
RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

WORKDIR /root/
COPY --from=build /node/root ./
WORKDIR /root/.simapp
RUN sed -i 's/127.0.0.1/0.0.0.0/g' ./config/config.toml
RUN chmod -R 0777 ./

