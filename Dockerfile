# This docker image is designed to be used in CI for benchmarks and also by developers wanting an environment that always has the lastest rocksdb and cleveldb. 
# Since the SDK is not simapp, this Dockerfile doesn't build simapp, but instead provides a full development environment for the Cosmos SDK.
# Other Docker images, like simapp, or the various protobuf build iamges could be based on this one

FROM faddat/archlinux

RUN pacman -Syyu --noconfirm leveldb rocksdb go base-devel git
