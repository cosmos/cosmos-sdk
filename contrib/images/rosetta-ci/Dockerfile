FROM faddat/archlinux

RUN pacman -Syyu --noconfirm leveldb rocksdb go base-devel git python

# prepare node data
WORKDIR /node
COPY ./contrib/images/rosetta-ci/data.tar.gz data.tar.gz
RUN tar -zxvf data.tar.gz -C .

# build simd
WORKDIR /simd
COPY . ./
RUN go build -o simd ./simapp/simd/

RUN mv /simd/simd /bin/simd

WORKDIR /rosetta
COPY ./contrib/rosetta/configuration ./
RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

RUN mv /node/root /root/
WORKDIR /root/.simapp

RUN chmod -R 0777 ./
