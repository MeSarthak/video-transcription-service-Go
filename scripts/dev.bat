@echo off
echo =======================================================
echo TranscribeX - Distributed Video Transcription Engine
echo Local Development Launcher
echo =======================================================

echo Checking Go version...
go version
if %ERRORLEVEL% NEQ 0 (
    echo Error: Go is not installed or not in PATH.
    exit /b 1
)

echo Building API server...
go build -o bin/api.exe ./cmd/api
if %ERRORLEVEL% NEQ 0 (
    echo Error building API server.
    exit /b 1
)

echo Building Worker daemon...
go build -o bin/worker.exe ./cmd/worker
if %ERRORLEVEL% NEQ 0 (
    echo Error building Worker daemon.
    exit /b 1
)

echo.
echo =======================================================
echo Both binaries compiled successfully!
echo To run locally:
echo 1. Start API:    .\bin\api.exe
echo 2. Start Worker: .\bin\worker.exe
echo 3. Start UI:     cd frontend ^&^& npm run dev
echo =======================================================
