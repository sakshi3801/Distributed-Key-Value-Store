package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/kvstore/distributed-kv/api"
	"github.com/kvstore/distributed-kv/internal/hash"
	"github.com/kvstore/distributed-kv/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type member struct {
	Address  string
	LastSeen time.Time
	Term     int64
}

type Node struct {
	ID       string
	Address  string
	Store    *store.Store
	Ring     *hash.Ring
	Replicas int

	mu        sync.RWMutex
	members   map[string]*member
	term      int64
	leaderID  string
	votedFor  string
	conns     map[string]*grpc.ClientConn
}

func NewNode(id, address string, replicas int) *Node {
	r := hash.NewRing(64)
	n := &Node{
		ID:       id,
		Address:  address,
		Store:    store.New(),
		Ring:     r,
		Replicas: replicas,
		members:  make(map[string]*member),
		conns:    make(map[string]*grpc.ClientConn),
	}
	r.Add(id)
	return n
}

func (n *Node) Ring() *hash.Ring {
	return n.Ring
}

func (n *Node) Store() *store.Store {
	return n.Store
}

func (n *Node) GetConn(nodeID string) (*grpc.ClientConn, error) {
	n.mu.RLock()
	addr := ""
	if m, ok := n.members[nodeID]; ok {
		addr = m.Address
	}
	conn, ok := n.conns[nodeID]
	n.mu.RUnlock()
	if ok && conn != nil {
		return conn, nil
	}
	if addr == "" {
		return nil, fmt.Errorf("unknown node %s", nodeID)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if c, ok := n.conns[nodeID]; ok && c != nil {
		return c, nil
	}
	c, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	n.conns[nodeID] = c
	return c, nil
}

func (n *Node) UpdateMembers(ids, addresses []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, id := range ids {
		if id == n.ID() {
			continue
		}
		addr := ""
		if i < len(addresses) {
			addr = addresses[i]
		}
		if addr == "" {
			continue
		}
		n.members[id] = &member{Address: addr, LastSeen: time.Now(), Term: n.term}
		if !n.Ring.Contains(id) {
			n.Ring.Add(id)
		}
	}
	for id := range n.members {
		found := false
		for _, x := range ids {
			if x == id {
				found = true
				break
			}
		}
		if !found {
			delete(n.members, id)
			delete(n.conns, id)
			n.Ring.Remove(id)
		}
	}
}

func (n *Node) ReplicateToReplicas(ctx context.Context, key string, value []byte, version int64, isDelete bool) {
	nodes := n.Ring.GetN(key, n.Replicas)
	for _, nodeID := range nodes {
		if nodeID == n.ID() {
			continue
		}
		conn, err := n.GetConn(nodeID)
		if err != nil {
			log.Printf("replicate to %s: %v", nodeID, err)
			continue
		}
		client := api.NewKeyValueClient(conn)
		_, _ = client.Replicate(ctx, &api.ReplicateRequest{
			Key:      key,
			Value:    value,
			Version:  version,
			IsDelete: isDelete,
		})
	}
}

func (n *Node) StartLeaderElection(ctx context.Context) bool {
	n.mu.Lock()
	n.term++
	term := n.term
	n.votedFor = n.ID()
	n.leaderID = ""
	n.mu.Unlock()

	votes := 1
	for id := range n.members {
		conn, err := n.GetConn(id)
		if err != nil {
			continue
		}
		client := api.NewClusterClient(conn)
		resp, err := client.RequestVote(ctx, &api.RequestVoteRequest{
			NodeId:       n.ID(),
			Term:         term,
			LastLogIndex: n.Store.Version(),
		})
		if err != nil {
			continue
		}
		if resp.VoteGranted {
			votes++
		}
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	quorum := (len(n.members)+1)/2 + 1
	if votes >= quorum {
		n.leaderID = n.ID()
		return true
	}
	return false
}

func (n *Node) IsLeader() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.leaderID == n.ID()
}

func (n *Node) SendHeartbeats(ctx context.Context) {
	n.mu.RLock()
	term := n.term
	members := make(map[string]string)
	for id, m := range n.members {
		members[id] = m.Address
	}
	n.mu.RUnlock()
	for id, addr := range members {
		conn, err := n.GetConn(id)
		if err != nil {
			continue
		}
		client := api.NewClusterClient(conn)
		_, _ = client.Heartbeat(ctx, &api.HeartbeatRequest{
			NodeId:   n.ID(),
			Address:  n.Address,
			Term:     term,
			IsLeader: true,
		})
		_ = addr
	}
}

func (n *Node) SeedPeers(peerList []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, p := range peerList {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id := p
		addr := p
		if idx := strings.Index(p, "="); idx > 0 {
			id = strings.TrimSpace(p[:idx])
			addr = strings.TrimSpace(p[idx+1:])
		}
		if id == n.ID() {
			continue
		}
		n.members[id] = &member{Address: addr, LastSeen: time.Now(), Term: n.term}
		if !n.Ring.Contains(id) {
			n.Ring.Add(id)
		}
	}
}

func (n *Node) SetLeader(id string, term int64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if term >= n.term {
		n.term = term
		n.leaderID = id
	}
}
