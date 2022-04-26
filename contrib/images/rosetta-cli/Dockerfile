FROM faddat/archlinux

RUN pacman -Syyu --noconfirm leveldb rocksdb go base-devel git python

ARG ROSETTA_VERSION="v0.7.3"

# build rosetta CLI
WORKDIR /rosetta
RUN git clone https://github.com/coinbase/rosetta-cli .
RUN git checkout tags/$ROSETTA_VERSION
RUN go build -o rosetta-cli ./main.go

RUN cp /rosetta/rosetta-cli /bin/rosetta-cli
