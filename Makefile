ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: proto build run run-config test clean

proto:
	@echo "generating proto files..."
	@mkdir -p gen/api
	@protoc \
     		-I api \
     		-I . \
     		--go_out=gen/api --go_opt=paths=source_relative \
     		--go-grpc_out=gen/api --go-grpc_opt=paths=source_relative \
			--grpc-gateway_out=gen/api --grpc-gateway_opt=paths=source_relative \
     		api/*.proto
	@echo "proto files generated successfully."

build:
	@echo "building the application..."
	@go build -o ./bin/order-manager ./cmd/order-manager/main.go

run: build
	@echo "running the application..."
	@./bin/order-manager

run-config: build
	@echo "running the application with a config file..."
	@./bin/order-manager --config=./config/settings.yml

test:
	@echo "running tests..."
	@go test ./tests

clean:
	@echo "cleaning up..."
	@rm -rf ./gen ./bin
