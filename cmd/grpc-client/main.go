package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pt "gitlab.com/marsskom/burro/internal/proto"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:7777",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pt.NewBurroClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.Subscribe(ctx, &pt.SubscribeRequest{})
	if err != nil {
		log.Fatal("subscribe error:", err)
	}

	log.Println("listening for events...")

	for {
		event, err := stream.Recv()
		if err != nil {
			log.Fatal("recv error:", err)
		}

		log.Printf("EVENT: %+v", event)
	}
}
