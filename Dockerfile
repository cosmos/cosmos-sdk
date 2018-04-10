# Simple usage with a mounted data directory:
# > docker build -t gaia .
# > docker run -v $HOME/.gaiad:/root/.gaiad gaia init
# > docker run -v $HOME/.gaiad:/root/.gaiad gaia start

FROM alpine:edge

# Install minimum necessary dependencies

ENV PACKAGES go glide make git libc-dev bash
RUN apk add --no-cache $PACKAGES

# Set up GOPATH & PATH

ENV GOPATH       /root/go
ENV BASE_PATH    $GOPATH/src/github.com/cosmos
ENV REPO_PATH    $BASE_PATH/cosmos-sdk
ENV WORKDIR      /cosmos/
ENV PATH         $GOPATH/bin:$PATH

# Link expected Go repo path

RUN mkdir -p $WORKDIR $GOPATH/pkg $ $GOPATH/bin $BASE_PATH && ln -sf $WORKDIR $REPO_PATH

# Add source files

ADD . $WORKDIR

# Build cosmos-sdk

RUN cd $REPO_PATH && make get_tools && make get_vendor_deps && make all && make install

# Remove packages

RUN apk del $PACKAGES

# Set entrypoint

ENTRYPOINT ["/root/go/bin/gaiad"]
