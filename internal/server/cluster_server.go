package server

import (
	"context"
	"time"

	"github.com/kvstore/distributed-kv/api"
)

// ClusterServer implements the Cluster gRPC service for membership and leader election.
type ClusterServer struct {
	api.UnimplementedClusterServer
	node *Node
}

// NewClusterServer returns a new ClusterServer.
func NewClusterServer(node *Node) *ClusterServer {
	return &ClusterServer{node: node}
}

// Heartbeat is called by the leader to maintain authority, or by nodes to register.
func (s *ClusterServer) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.HeartbeatResponse, error) {
	if req == nil {
		return &api.HeartbeatResponse{Accepted: false}, nil
	}
	s.node.mu.Lock()
	defer s.node.mu.Unlock()
	// Update or add node in membership.
	s.node.members[req.NodeId] = &member{Address: req.Address, LastSeen: time.Now(), Term: req.Term}
	if req.IsLeader && req.Term >= s.node.term {
		s.node.leaderID = req.NodeId
		s.node.term = req.Term
	}
	return &api.HeartbeatResponse{Term: s.node.term, Accepted: true}, nil
}

// RequestVote is used for leader election (simple majority).
func (s *ClusterServer) RequestVote(ctx context.Context, req *api.RequestVoteRequest) (*api.RequestVoteResponse, error) {
	if req == nil {
		return &api.RequestVoteResponse{VoteGranted: false}, nil
	}
	s.node.mu.Lock()
	defer s.node.mu.Unlock()
	resp := &api.RequestVoteResponse{Term: s.node.term}
	if req.Term >= s.node.term {
		s.node.term = req.Term
		s.node.votedFor = req.NodeId
		resp.VoteGranted = true
	}
	return resp, nil
}

// GetClusterState returns current cluster view (nodes, leader, term).
func (s *ClusterServer) GetClusterState(ctx context.Context, req *api.GetClusterStateRequest) (*api.GetClusterStateResponse, error) {
	s.node.mu.RLock()
	defer s.node.mu.RUnlock()
	ids := make([]string, 0, len(s.node.members))
	addrs := make([]string, 0, len(s.node.members))
	for id, m := range s.node.members {
		ids = append(ids, id)
		addrs = append(addrs, m.Address)
	}
	return &api.GetClusterStateResponse{
		NodeIds:   ids,
		Addresses: addrs,
		LeaderId:  s.node.leaderID,
		Term:      s.node.term,
	}, nil
}
