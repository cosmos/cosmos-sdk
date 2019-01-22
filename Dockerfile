# Simple usage with a mounted data directory:
# > docker build -t gaia .
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.gaiad:/root/.gaiad -v ~/.gaiacli:/root/.gaiacli gaia gaiad init
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.gaiad:/root/.gaiad -v ~/.gaiacli:/root/.gaiacli gaia gaiad start
FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES make git libc-dev bash gcc linux-headers eudev-dev

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/cosmos-sdk

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make tools && \
    make vendor-deps && \
    make build && \
    make install

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/gaiad /usr/bin/gaiad
COPY --from=build-env /go/bin/gaiacli /usr/bin/gaiacli

# Run gaiad by default, omit entrypoint to ease using container with gaiacli
CMD ["gaiad"]
