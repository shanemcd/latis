// latis is the CLI and control plane for managing distributed AI agents.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	latisv1 "github.com/shanemcd/latis/gen/go/latis/v1"
	quictransport "github.com/shanemcd/latis/pkg/transport/quic"
)

var (
	addr    = flag.String("addr", "localhost:4433", "unit address to connect to")
	prompt  = flag.String("prompt", "", "prompt to send (if empty, sends ping)")
)

func main() {
	flag.Parse()

	log.Printf("latis connecting to %s", *addr)

	// TLS config for client (skip verification for development)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"latis"},
	}

	// Create gRPC connection over QUIC
	dialer := quictransport.NewDialer(tlsConfig, nil)

	conn, err := grpc.NewClient(
		*addr,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := latisv1.NewLatisServiceClient(conn)

	// Establish bidirectional stream
	ctx := context.Background()
	stream, err := client.Connect(ctx)
	if err != nil {
		log.Fatalf("failed to establish stream: %v", err)
	}

	// Send message based on flags
	msgID := uuid.New().String()

	if *prompt != "" {
		// Send prompt
		log.Printf("sending prompt: %s", *prompt)
		if err := stream.Send(&latisv1.ConnectRequest{
			Id: msgID,
			Payload: &latisv1.ConnectRequest_PromptSend{
				PromptSend: &latisv1.PromptSend{
					Content: *prompt,
				},
			},
		}); err != nil {
			log.Fatalf("failed to send prompt: %v", err)
		}
	} else {
		// Send ping
		log.Println("sending ping")
		if err := stream.Send(&latisv1.ConnectRequest{
			Id: msgID,
			Payload: &latisv1.ConnectRequest_Ping{
				Ping: &latisv1.Ping{
					Timestamp: time.Now().UnixNano(),
				},
			},
		}); err != nil {
			log.Fatalf("failed to send ping: %v", err)
		}
	}

	// Receive responses
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed")
			break
		}
		if err != nil {
			log.Fatalf("failed to receive: %v", err)
		}

		// Handle response based on type
		switch payload := resp.Payload.(type) {
		case *latisv1.ConnectResponse_ResponseChunk:
			fmt.Print(payload.ResponseChunk.Content)

		case *latisv1.ConnectResponse_ResponseComplete:
			fmt.Println()
			log.Printf("response complete for request %s", payload.ResponseComplete.RequestId)
			os.Exit(0)

		case *latisv1.ConnectResponse_Pong:
			latency := time.Now().UnixNano() - payload.Pong.PingTimestamp
			log.Printf("pong received, latency=%v", time.Duration(latency))
			os.Exit(0)

		case *latisv1.ConnectResponse_Error:
			log.Printf("error: %s - %s", payload.Error.Code, payload.Error.Message)
			os.Exit(1)

		default:
			log.Printf("unhandled response type: %T", payload)
		}
	}
}
