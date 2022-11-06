FROM golang:1.19-alpine as build

RUN apk add --no-cache tar git

# prepare node data
WORKDIR /node
COPY ./contrib/rosetta/rosetta-ci/data.tar.gz data.tar.gz
RUN tar -zxvf data.tar.gz -C .

# build simd
WORKDIR /simd
COPY . ./
RUN cd simapp && go build -o simd ./simd/

FROM alpine
RUN apk add gcc git libc-dev python3 --no-cache

ENV PATH=$PATH:/bin

COPY --from=build /simd/simapp/simd/simd /bin/simd

WORKDIR /rosetta
COPY ./contrib/rosetta/configuration ./

# Set GENESIS_HASH env needed for api correctness
ENV GENESIS_HASH "01331100220a94b9acc0ceef697a6db44ba70fc61e9e6c24802974a0b6cba29b"

RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

COPY --from=build /node/root /root/
WORKDIR /root/.simapp

RUN chmod -R 0777 ./
