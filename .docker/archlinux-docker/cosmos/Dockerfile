# This dockerfile will be used to bootstrap the docker changes for the cosmos SDK, and will be built hourly.
# We'll also move ghcr.io/notional-labs/archlinux (this will be ghcr.io/cosmos/cosmos-sdk post merge) to an hourly schedule, ensuring that when the images are put to use, they dont' need to update as many packages.

FROM ghcr.io/notional-labs/archlinux

RUN pacman -Syyu --noconfirm go base-devel git leveldb rocksdb make python snappy lz4
