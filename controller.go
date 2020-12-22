package main

import (
	"context"
	"log"
	"time"

	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	_ "github.com/VertexC/log-formatter/proto/pkg/controller"
	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:2001", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := agentpb.NewLogFormatterAgentClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	heartbeatRequest := &agentpb.HeartBeatRequest{}
	r, err := c.GetHeartBeat(ctx, heartbeatRequest)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Got Heartbeat: %+v\n", *r)
}
