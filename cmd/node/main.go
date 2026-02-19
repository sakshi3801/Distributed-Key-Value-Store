package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kvstore/distributed-kv/api"
	"github.com/kvstore/distributed-kv/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	id := flag.String("id", "node1", "Node ID")
	addr := flag.String("addr", "0.0.0.0:50051", "Listen address")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses (host:port) for discovery; each can be id=host:port")
	replicas := flag.Int("replicas", 2, "Replication factor (number of nodes that hold each key)")
	flag.Parse()

	node := server.NewNode(*id, *addr, *replicas)

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	api.RegisterKeyValueServer(srv, server.NewKVServer(node))
	api.RegisterClusterServer(srv, server.NewClusterServer(node))

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *peers != "" {
		peerList := strings.Split(*peers, ",")
		node.SeedPeers(peerList)
		go runDiscovery(ctx, node, *peers)
		go runElectionAndHeartbeat(ctx, node)
	}

	log.Printf("node %s listening on %s", *id, *addr)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	srv.GracefulStop()
}

// runDiscovery periodically fetches cluster state from peers and updates the ring.
func runDiscovery(ctx context.Context, node *server.Node, peersStr string) {
	peerAddrs := strings.Split(peersStr, ",")
	var conns []*grpc.ClientConn
	for _, p := range peerAddrs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		addr := p
		if idx := strings.Index(p, "="); idx > 0 {
			addr = strings.TrimSpace(p[idx+1:])
		}
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("dial peer %s: %v", addr, err)
			continue
		}
		conns = append(conns, conn)
	}
	if len(conns) == 0 {
		return
	}
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			for _, c := range conns {
				_ = c.Close()
			}
			return
		case <-tick.C:
			var lastErr error
			for _, conn := range conns {
				client := api.NewClusterClient(conn)
				st, err := client.GetClusterState(ctx, &api.GetClusterStateRequest{})
				if err != nil {
					lastErr = err
					continue
				}
				node.UpdateMembers(st.NodeIds, st.Addresses)
				break
			}
			_ = lastErr
		}
	}
}

// runElectionAndHeartbeat runs leader election and, if we become leader, sends heartbeats.
func runElectionAndHeartbeat(ctx context.Context, node *server.Node) {
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if node.IsLeader() {
				node.SendHeartbeats(ctx)
			} else {
				// Try to become leader
				if node.StartLeaderElection(ctx) {
					log.Printf("node %s became leader", node.ID)
				}
			}
		}
	}
}
