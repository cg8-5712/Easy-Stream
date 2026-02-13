.PHONY: help build build-frontend build-backend build-backend-only clean dev install

help:
	@echo "Available commands:"
	@echo "  make install           - Install frontend dependencies"
	@echo "  make build-frontend    - Build frontend only"
	@echo "  make build-backend-only- Build backend only (no frontend)"
	@echo "  make build-backend     - Build backend with embedded frontend"
	@echo "  make build             - Build both frontend and backend (with embedding)"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make dev               - Run in development mode"

install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

build-frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build

build-backend-only:
	@echo "Building backend only (no frontend)..."
	go build -o easy-stream.exe ./cmd/server

build-backend:
	@echo "Building backend with embedded frontend..."
	@echo "Copying frontend dist to web..."
	@rm -rf web/dist
	@cp -r frontend/dist web/dist
	go build -tags embed_frontend -o easy-stream.exe ./cmd/server

build: build-frontend build-backend
	@echo "Build complete!"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf web/dist
	rm -f easy-stream.exe
	rm -f easy-stream

dev:
	@echo "Running in development mode..."
	go run ./cmd/server
