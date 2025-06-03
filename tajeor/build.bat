@echo off
echo Building Tajeor Blockchain...

:: Clean previous builds
if exist "build" rmdir /s /q build
mkdir build

:: Build the binary
go build -o build/tajeord.exe ./cmd/tajeord

:: Check if build was successful
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

echo Build successful!
echo Binary created at: build\tajeord.exe 