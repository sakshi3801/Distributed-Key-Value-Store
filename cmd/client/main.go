package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kvstore/distributed-kv/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "Node address")
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: client get <key> | set <key> <value>\n")
		os.Exit(1)
	}
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	client := api.NewKeyValueClient(conn)
	ctx := context.Background()
	switch args[0] {
	case "get":
		resp, err := client.Get(ctx, &api.GetRequest{Key: args[1]})
		if err != nil {
			log.Fatalf("get: %v", err)
		}
		if !resp.Found {
			fmt.Println("(not found)")
			return
		}
		// Print as string for readability; use -raw to get base64
		if len(args) > 2 && args[2] == "-raw" {
			fmt.Println(base64.StdEncoding.EncodeToString(resp.Value))
		} else {
			fmt.Println(string(resp.Value))
		}
	case "set":
		val := ""
		if len(args) > 2 {
			val = args[2]
		}
		_, err := client.Set(ctx, &api.SetRequest{Key: args[1], Value: []byte(val)})
		if err != nil {
			log.Fatalf("set: %v", err)
		}
		fmt.Println("ok")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		os.Exit(1)
	}
}
