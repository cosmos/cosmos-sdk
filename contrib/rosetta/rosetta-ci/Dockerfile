FROM golang:1.19-alpine as build

RUN apk add --no-cache tar git

# prepare node data
WORKDIR /node
COPY ./contrib/rosetta/rosetta-ci/data.tar.gz data.tar.gz
RUN tar -zxvf data.tar.gz -C .

# build simd
WORKDIR /simd
COPY . ./
RUN go build -o simd ./simapp/simd/

FROM alpine
RUN apk add gcc git libc-dev python3 --no-cache

ENV PATH=$PATH:/bin

COPY --from=build /simd/simd /bin/simd

WORKDIR /rosetta
COPY ./contrib/rosetta/configuration ./
RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

COPY --from=build /node/root /root/
WORKDIR /root/.simapp

RUN chmod -R 0777 ./
