#!/usr/bin/env bash
set -e

echo "======================================================="
echo "TranscribeX - Distributed Video Transcription Engine"
echo "Local Development Launcher"
echo "======================================================="

echo "Building API server..."
go build -o bin/api ./cmd/api

echo "Building Worker daemon..."
go build -o bin/worker ./cmd/worker

echo "Compilation successful!"
echo "Run options:"
echo "  1) ./bin/api"
echo "  2) ./bin/worker"
echo "  3) cd frontend && npm run dev"
echo "  4) docker-compose up -d --build"
