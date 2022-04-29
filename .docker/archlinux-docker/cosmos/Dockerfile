# This dockerfile will be used to bootstrap the docker changes for the cosmos SDK, and will be built hourly.
# We'll also move faddat/archlinux to an hourly schedule, ensuring that when the images are put to use, they dont' need to update as many packages.

FROM ghcr.io/faddat/archlinux

RUN pacman -Syyu --noconfirm go base-devel git leveldb rocksdb make python snappy lz4
