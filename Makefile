.PHONY: proto build run test clean

proto:
	@echo "generating proto files..."
	@mkdir -p gen
	@protoc \
     		-I api \
     		--go_out=api/gen --go_opt=paths=source_relative \
     		--go-grpc_out=api/gen --go-grpc_opt=paths=source_relative \
     		api/*.proto
	@echo "proto files generated successfully."

build:
	@echo "building the application..."
	@go build -o ./bin/order-manager ./cmd/order-manager/main.go

run: build
	@echo "running the application..."
	@./bin/sso --config=./config/settings.yaml

test:
	@echo "running tests..."
	@go test ./tests

clean:
	@echo "cleaning up..."
	@rm -rf ./gen ./bin
