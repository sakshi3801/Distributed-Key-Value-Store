# Distributed Key-Value Store (gRPC + Go)

A distributed key-value store with **multi-node replication**, **gRPC** communication, **leader election**, **fault tolerance**, and **consistent hashing** for data partitioning. Supports dynamic node addition and failure recovery, and can be run as a simulated cluster with Docker.

## Features

- **gRPC API**: Get, Set, Delete, and replication RPCs between nodes
- **Consistent hashing**: Keys are partitioned across nodes with minimal movement on topology changes
- **Replication**: Configurable replication factor; writes are replicated to N nodes per key
- **Leader election**: Bully-style election; leader sends heartbeats to maintain authority
- **Cluster membership**: Heartbeat and GetClusterState for discovery and ring updates
- **Fault tolerance**: Node failures are detected via discovery; ring is updated and data remains available from replicas
- **Docker**: Multi-node cluster via `docker compose`

## Prerequisites

- **Go 1.21+** (for local build)
- **Docker & Docker Compose** (for containerized cluster)
- Optional: **protoc**, **protoc-gen-go**, **protoc-gen-go-grpc** (to regenerate `api/` from `api/kv.proto`)

## Quick Start (Docker)

Run a 3-node cluster:

```bash
docker compose up --build -d
```

Nodes listen on:

- `node1`: `localhost:50051`
- `node2`: `localhost:50052`
- `node3`: `localhost:50053`

Use any node as the gRPC endpoint; requests are forwarded to the correct node by consistent hashing.

## Local Build and Run

1. **Install tools** (optional, only if you change `.proto`):

   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   make proto
   ```

2. **Build and run a single node**:

   ```bash
   go build -o bin/node ./cmd/node
   ./bin/node -id node1 -addr 0.0.0.0:50051
   ```

3. **Run a 3-node cluster locally** (in 3 terminals):

   ```bash
   ./bin/node -id node1 -addr 0.0.0.0:50051 -peers "node2=127.0.0.1:50052,node3=127.0.0.1:50053" -replicas 2
   ./bin/node -id node2 -addr 0.0.0.0:50052 -peers "node1=127.0.0.1:50051,node3=127.0.0.1:50053" -replicas 2
   ./bin/node -id node3 -addr 0.0.0.0:50053 -peers "node1=127.0.0.1:50051,node2=127.0.0.1:50052" -replicas 2
   ```

## Flags

| Flag       | Default   | Description                                              |
|-----------|-----------|----------------------------------------------------------|
| `-id`     | `node1`   | Unique node ID                                          |
| `-addr`   | `:50051`  | Listen address (host:port)                               |
| `-peers`  | (empty)   | Comma-separated peers: `addr` or `id=host:port`         |
| `-replicas` | `2`     | Replication factor (number of nodes storing each key)   |

## Architecture

- **Consistent hashing** (`internal/hash`): Partitions keys across nodes; adding/removing nodes causes only local key movement.
- **Store** (`internal/store`): In-memory KV with versioning for ordering and replication.
- **Server** (`internal/server`): Implements `KeyValue` and `Cluster` gRPC services; forwards Get/Set/Delete to the ring owner; replicates writes to N nodes.
- **Node** (`internal/server/node.go`): Holds ring, store, membership, leader state; runs election and heartbeats when peers are configured.
