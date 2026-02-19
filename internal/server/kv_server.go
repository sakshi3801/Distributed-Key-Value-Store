package server

import (
	"context"
	"log"

	"github.com/kvstore/distributed-kv/api"
)

// KVServer implements the KeyValue gRPC service.
type KVServer struct {
	api.UnimplementedKeyValueServer
	node *Node
}

// NewKVServer returns a new KVServer backed by the given node.
func NewKVServer(node *Node) *KVServer {
	return &KVServer{node: node}
}

// Get returns the value for the key from the local store or forwards to the responsible node.
func (s *KVServer) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	if req == nil || req.Key == "" {
		return &api.GetResponse{Found: false}, nil
	}
	// If we are the owner (according to consistent hash), serve from local store.
	owner := s.node.Ring().Get(req.Key)
	if owner == s.node.ID() {
		val, _, found := s.node.Store().Get(req.Key)
		return &api.GetResponse{Value: val, Found: found}, nil
	}
	// Forward to the owner node.
	if owner != "" {
		conn, err := s.node.GetConn(owner)
		if err != nil {
			log.Printf("get: forward to %s: %v", owner, err)
			return &api.GetResponse{Found: false}, nil
		}
		client := api.NewKeyValueClient(conn)
		return client.Get(ctx, req)
	}
	return &api.GetResponse{Found: false}, nil
}

// Set writes the key to the responsible node and replicates.
func (s *KVServer) Set(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {
	if req == nil || req.Key == "" {
		return &api.SetResponse{Ok: false}, nil
	}
	owner := s.node.Ring().Get(req.Key)
	if owner == "" {
		return &api.SetResponse{Ok: false}, nil
	}
	// If we are the owner, apply locally and replicate.
	if owner == s.node.ID() {
		var version int64
		if req.Version > 0 {
			version = req.Version
			s.node.Store().SetVersion(req.Key, req.Value, version)
		} else {
			version = s.node.Store().Set(req.Key, req.Value)
		}
		s.node.ReplicateToReplicas(ctx, req.Key, req.Value, version, false)
		return &api.SetResponse{Ok: true}, nil
	}
	// Forward to owner.
	conn, err := s.node.GetConn(owner)
	if err != nil {
		log.Printf("set: forward to %s: %v", owner, err)
		return &api.SetResponse{Ok: false}, nil
	}
	client := api.NewKeyValueClient(conn)
	return client.Set(ctx, req)
}

// Delete removes the key from the responsible node and replicates.
func (s *KVServer) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	if req == nil || req.Key == "" {
		return &api.DeleteResponse{Ok: false}, nil
	}
	owner := s.node.Ring().Get(req.Key)
	if owner == "" {
		return &api.DeleteResponse{Ok: false}, nil
	}
	if owner == s.node.ID() {
		var version int64
		if req.Version > 0 {
			version = req.Version
			s.node.Store().DeleteVersion(req.Key, version)
		} else {
			version = s.node.Store().Delete(req.Key)
		}
		s.node.ReplicateToReplicas(ctx, req.Key, nil, version, true)
		return &api.DeleteResponse{Ok: true}, nil
	}
	conn, err := s.node.GetConn(owner)
	if err != nil {
		return &api.DeleteResponse{Ok: false}, nil
	}
	client := api.NewKeyValueClient(conn)
	return client.Delete(ctx, req)
}

// Replicate applies a replicated write from the leader (no forwarding).
func (s *KVServer) Replicate(ctx context.Context, req *api.ReplicateRequest) (*api.ReplicateResponse, error) {
	if req == nil || req.Key == "" {
		return &api.ReplicateResponse{Ok: false}, nil
	}
	if req.IsDelete {
		ok := s.node.Store().DeleteVersion(req.Key, req.Version)
		return &api.ReplicateResponse{Ok: ok}, nil
	}
	ok := s.node.Store().SetVersion(req.Key, req.Value, req.Version)
	return &api.ReplicateResponse{Ok: ok}, nil
}
