# This is the Cosmos-SDK Development environment as a container.
# It contains everything needed to use the Cosmos-SDK to its fullest.
# It does not contain the cosmos sdk itself, because it is meant to be called from Github actions or other CI systems.
# This should be usable as the build environment for any cosmos chain, and includes Rust due to the growing popularity of CosmWasm.
# This container is built daily using a cron job on the cosmos-sdk repository, and you can pull it like: docker pull ghcr.io/cosmos/cosmos-sdk
FROM faddat/archlinux

RUN pacman -Syyu --noconfirm go base-devel python3 git rocksdb leveldb
