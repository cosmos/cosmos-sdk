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

# Ensure required tools are installed
command -v docker >/dev/null 2>&1 || { echo >&2 "docker is required but it's not installed. Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required but it's not installed. Aborting."; exit 1; }

# Ensure Docker daemon is running
if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is not running. Please start Docker and try again."
    exit 1
fi

# Cleanup Function
cleanup() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Cleaning up any existing container with the same name..."
    if docker rm -f ${CONTAINER_NAME} 2>/dev/null; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] Container ${CONTAINER_NAME} successfully removed."
    else
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] No container named ${CONTAINER_NAME} found or removal failed."
    fi
}

# Build Image Function with retry logic
build_image() {
    local image_name="$1"
    local dockerfile_path="$2"
    local clangfile_path="$3"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Building Docker image: ${image_name}..."
    for i in {1..3}; do
        if docker build -t ${image_name} -f ${dockerfile_path} --build-arg CLANGFILE=${clangfile_path} ${CONTEXT_DIR} | tee -a ${LOG_FILE}; then
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] Docker image ${image_name} built successfully."
            return 0
        else
            echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed. Retry $i/3."
            sleep 5
        fi
    done
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: Docker image build for ${image_name} failed after 3 attempts. Check ${LOG_FILE} for details."
    exit 1
}

# Run Container Function
run_container() {
    local image_name="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Running Docker container: ${CONTAINER_NAME} for ${image_name}..."
    if docker run -it --rm \
        --name ${CONTAINER_NAME} \
        --mount type=bind,source=${WORKSPACE_DIR},target=/workspace \
        --network none \
        ${image_name}; then
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

# Trap exit signals to clean up
trap cleanup EXIT

# Main Execution Workflow
echo "[$(date +'%Y-%m-%d %H:%M:%S')] Starting Runner Image Setup for Ubuntu 24.04 and Windows Server 2025 with Clang configurations..." | tee -a ${LOG_FILE}

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

echo "[$(date +'%Y-%m-%d %H:%M:%S')] Runner Image Setup for both Ubuntu 24.04 and Windows Server 2025 with Clang configurations completed." | tee -a ${LOG_FILE}
