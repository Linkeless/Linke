.PHONY: build run test clean swagger dev

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ docs/

swagger:
	swag init -g cmd/server/main.go -o docs

dev: swagger
	go run cmd/server/main.go

install:
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  swagger  - Generate Swagger documentation"
	@echo "  dev      - Run in development mode with swagger"
	@echo "  install  - Install dependencies"