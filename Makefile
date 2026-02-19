.PHONY: proto proto-docker build run test docker-build docker-up docker-down

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/kv.proto

# Generate proto without local protoc (requires Docker)
proto-docker:
	docker run --rm -v "$$(pwd):/workspace" -w /workspace \
	  golang:1.21-alpine sh -c 'apk add --no-cache protobuf-dev protobuf && \
	  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
	  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/kv.proto'

build: proto
	go build -o bin/node ./cmd/node

run: build
	./bin/node

test:
	go test ./...

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
