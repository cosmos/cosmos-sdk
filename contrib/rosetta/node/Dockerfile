FROM golang:1.15-alpine as build

RUN apk add --no-cache tar

# prepare node data
WORKDIR /node
COPY ./contrib/rosetta/node/data.tar.gz data.tar.gz
RUN tar -zxvf data.tar.gz -C .

# build simd and simcli
WORKDIR /simd
COPY . ./
RUN go build -o simd ./contrib/rosetta/simd/cmd/simd/simd.go
RUN go build -o simcli ./contrib/rosetta/simd/cmd/simcli/simcli.go

FROM alpine
RUN apk add gcc libc-dev python3 --no-cache

ENV PATH=$PATH:/bin

COPY --from=build /simd/simd /bin/simd
COPY --from=build /simd/simcli /bin/simcli

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

