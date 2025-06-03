# Setup script for Tajeor Blockchain
Write-Host "Setting up Tajeor Blockchain development environment..."

# Check if Go is installed
$goVersion = go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "Go is not installed. Please install Go 1.21 or higher from https://golang.org/dl/"
    exit 1
}

# Install required tools
Write-Host "Installing required tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/bufbuild/buf/cmd/buf@latest

# Install dependencies
Write-Host "Installing dependencies..."
go mod tidy
go mod download

# Check if build directory exists
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build"
}

Write-Host "Setup complete!"
Write-Host "You can now build the blockchain by running build.bat" 