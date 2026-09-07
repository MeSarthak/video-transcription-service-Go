.PHONY: all build build-api build-worker build-frontend test lint run-api run-worker run-frontend docker-up docker-down clean help

all: test build

help:
	@echo "TranscribeX - Distributed Video Transcription Engine"
	@echo ""
	@echo "Usage:"
	@echo "  make build         Build all Go binaries and Frontend"
	@echo "  make test          Run all Go unit & integration tests"
	@echo "  make run-api       Start the Go HTTP API Server"
	@echo "  make run-worker    Start the Go Background Worker"
	@echo "  make run-frontend  Start the React HeroUI dev server"
	@echo "  make docker-up     Start entire stack in Docker Compose"
	@echo "  make docker-down   Stop Docker Compose containers"
	@echo "  make clean         Remove build artifacts"

build: build-api build-worker build-frontend

build-api:
	@echo "Building API server..."
	go build -o bin/api ./cmd/api

build-worker:
	@echo "Building Worker daemon..."
	go build -o bin/worker ./cmd/worker

build-frontend:
	@echo "Building Frontend distribution..."
	cd frontend && npm run build

test:
	@echo "Running backend test suite..."
	go test -v ./...

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

run-frontend:
	cd frontend && npm run dev

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

clean:
	rm -rf bin/
	rm -rf frontend/dist/
