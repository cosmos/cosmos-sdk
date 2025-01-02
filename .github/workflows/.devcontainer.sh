name: "Lint PR"

on:
  pull_request_target:
    types:
      - opened
      - edited
      - synchronize

permissions:
  contents: read

jobs:
  main:
    permissions:
      pull-requests: read # for amannn/action-semantic-pull-request to analyze PRs
      statuses: write # for amannn/action-semantic-pull-request to mark status of analyzed PR
    runs-on: ubuntu-latest
    steps:
      - uses: amannn/action-semantic-pull-request@v5.5.3
        id: lint_pr_title
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - uses: marocchino/sticky-pull-request-comment@v2
        # When the previous steps fails, the workflow would stop. By adding this
        # condition you can continue the execution with the populated error message.
        if: always() && (steps.lint_pr_title.outputs.error_message != null)
        with:
          header: pr-title-lint-error
          message: |
            Hey there and thank you for opening this pull request! 👋🏼

            We require pull request titles to follow the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/) and it looks like your proposed title needs to be adjusted.

            Details:

            ```
            ${{ steps.lint_pr_title.outputs.error_message }}
            ```

      # Delete a previous comment when the issue has been resolved
      - if: ${{ steps.lint_pr_title.outputs.error_message == null }}
        uses: marocchino/sticky-pull-request-comment@v2
        with:
          header: pr-title-lint-error
          delete: true

#!/bin/bash

# Script to build and run Runner Images for Ubuntu 24.04 and Windows Server 2025 debugging
# with Clang setup and Cosmos SDK integration.

# Variables
UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu" # Adjust if Dockerfile for Ubuntu is in a different location
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows" # Adjust if Dockerfile for Windows is in a different location
CONTEXT_DIR="." # Adjust if the context is a different directory
WORKSPACE_DIR="$(pwd)" # Current directory as the workspace
UBUNTU_CLANGFILE_PATH="clangfile.ubuntu.json"
WINDOWS_CLANGFILE_PATH="clangfile.windows.json"
LOG_FILE="runner-images-build.log"

# JSON File Paths
CHAIN_INFO_JSON="chain_info_mainnets.json"
IBC_INFO_JSON="ibc_info.json"
ASSET_LIST_JSON="asset_list_mainnets.json"
COSMWASM_MSGS_JSON="cosmwasm_json_msgs.json"
OSMOSIS_MSGS_JSON="osmosis_json_msgs.json"

# Functions

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
}

# Build Image Function
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Check ${LOG_FILE} for details."
        exit 1
    fi
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \
        ${image_name}
    if [ $? -eq 0 ]; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} for ${image_name} ran successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Failed to run container ${CONTAINER_NAME} for ${image_name}."
        exit 1
    fi
}

# Validate JSON Configurations
validate_json_files() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Validating JSON configurations..."
    for file in $CHAIN_INFO_JSON $IBC_INFO_JSON $ASSET_LIST_JSON $COSMWASM_MSGS_JSON $OSMOSIS_MSGS_JSON; do
        if jq empty $file >/dev/null 2>&1; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] $file is valid."
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $file is not valid JSON."
            exit 1
        fi
    done
}

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..."

# Validate JSON configurations
validate_json_files

# Clean up any previous runs
cleanup

# Build the Ubuntu Docker image with Clang configuration
build_image ${UBUNTU_IMAGE_NAME} ${UBUNTU_DOCKERFILE_PATH} ${UBUNTU_CLANGFILE_PATH}

# Run the Ubuntu container
run_container ${UBUNTU_IMAGE_NAME}

# Build the Windows Docker image with Clang configuration
build_image ${WINDOWS_IMAGE_NAME} ${WINDOWS_DOCKERFILE_PATH} ${WINDOWS_CLANGFILE_PATH}

# Run the Windows container
run_container ${WINDOWS_IMAGE_NAME}

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed."

# Use base images for C++ development
FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04 AS ubuntu-base

# Ubuntu Environment Setup
FROM ubuntu-base AS ubuntu-setup
ARG REINSTALL_CMAKE_VERSION_FROM_SOURCE="none"
COPY ./reinstall-cmake.sh /tmp/
RUN if [ "${REINSTALL_CMAKE_VERSION_FROM_SOURCE}" != "none" ]; then \
    chmod +x /tmp/reinstall-cmake.sh && /tmp/reinstall-cmake.sh ${REINSTALL_CMAKE_VERSION_FROM_SOURCE}; \
fi \
&& rm -f /tmp/reinstall-cmake.sh \
&& apt-get update && export DEBIAN_FRONTEND=noninteractive \
&& apt-get -y install --no-install-recommends \
    python3-pip \
    nodejs \
    npm \
    openjdk-17-jdk \
    gdb \
    valgrind \
    lsof \
    git \
    clang-18 \
    libstdc++-12-dev \
    glibc-source \
&& apt-get clean && rm -rf /var/lib/apt/lists/*

# Python setup
RUN python3 -m pip install --upgrade pip

# Node.js setup
RUN npm install -g yarn

# Install vcpkg if not already present
ENV VCPKG_INSTALLATION_ROOT=/vcpkg
RUN git clone https://github.com/microsoft/vcpkg.git $VCPKG_INSTALLATION_ROOT \
&& cd $VCPKG_INSTALLATION_ROOT \
&& ./bootstrap-vcpkg.sh

# Copy project files into the container
COPY . /workspace
WORKDIR /workspace
CMD ["bash"]

# Use base images for C++ development
FROM mcr.microsoft.com/dotnet/framework/sdk:4.8-windowsservercore-ltsc2022 AS windows-base

# Windows Environment Setup
FROM windows-base AS windows-setup
SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
RUN iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1')); \
    choco install -y \
    msys2 \
    cmake \
    clang \
    python \
    nodejs \
    git \
    jdk17 \
    visualstudio2022buildtools --package-parameters "--add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"

# Setup environment variables
ENV PATH="${PATH};C:\msys64\usr\bin;C:\Program Files\Git\cmd"

# Install vcpkg for Windows
RUN git clone https://github.com/microsoft/vcpkg.git C:\vcpkg \
&& cd C:\vcpkg \
&& .\bootstrap-vcpkg.bat

# Copy project files into the container
COPY . C:\workspace
WORKDIR C:\workspace
CMD ["powershell"]

#!/bin/bash

# Script to build and run Runner Images for Ubuntu 24.04 and Windows Server 2025 debugging
# with Clang setup

# Variables
UBUNTU_IMAGE_NAME="runner-images-ubuntu-24.04"
WINDOWS_IMAGE_NAME="runner-images-windows-2025"
CONTAINER_NAME="runner-images-container"
UBUNTU_DOCKERFILE_PATH="./Dockerfile.ubuntu" # Adjust if Dockerfile for Ubuntu is in a different location
WINDOWS_DOCKERFILE_PATH="./Dockerfile.windows" # Adjust if Dockerfile for Windows is in a different location
CONTEXT_DIR="." # Adjust if the context is a different directory
WORKSPACE_DIR="$(pwd)" # Current directory as the workspace
UBUNTU_CLANGFILE_PATH="clangfile.ubuntu.json"
WINDOWS_CLANGFILE_PATH="clangfile.windows.json"
LOG_FILE="runner-images-build.log"

# Functions

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
}

# Build Image Function
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Check ${LOG_FILE} for details."
        exit 1
    fi
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \ # Ensures no network access for isolation
        ${image_name}
    if [ $? -eq 0 ]; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} for ${image_name} ran successfully."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Failed to run container ${CONTAINER_NAME} for ${image_name}."
        exit 1
    fi
}

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..."

# Clean up any previous runs
cleanup

# Build the Ubuntu Docker image with Clang configuration
build_image ${UBUNTU_IMAGE_NAME} ${UBUNTU_DOCKERFILE_PATH} ${UBUNTU_CLANGFILE_PATH}

# Run the Ubuntu container
run_container ${UBUNTU_IMAGE_NAME}

# Build the Windows Docker image with Clang configuration
build_image ${WINDOWS_IMAGE_NAME} ${WINDOWS_DOCKERFILE_PATH} ${WINDOWS_CLANGFILE_PATH}

# Run the Windows container
run_container ${WINDOWS_IMAGE_NAME}

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed."

