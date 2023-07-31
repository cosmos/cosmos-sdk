FROM ghcr.io/cosmos/simapp:latest

# Set up dependencies
ENV PACKAGES curl make bash jq sed

# Install minimum necessary dependencies
RUN apk add --no-cache $PACKAGES

WORKDIR /opt
