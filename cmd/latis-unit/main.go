// latis-unit is the agent endpoint daemon that runs on remote machines.
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"time"

	"google.golang.org/grpc"

	latisv1 "github.com/shanemcd/latis/gen/go/latis/v1"
	quictransport "github.com/shanemcd/latis/pkg/transport/quic"
)

var (
	addr = flag.String("addr", "localhost:4433", "address to listen on")
)

func main() {
	flag.Parse()

	log.Printf("latis-unit starting on %s", *addr)

	// Generate a self-signed TLS certificate for development
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		log.Fatalf("failed to generate TLS config: %v", err)
	}

	// Create QUIC listener
	listener, err := quictransport.Listen(*addr, tlsConfig, nil)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("listening on %s (QUIC)", listener.Addr())

	// Create gRPC server
	grpcServer := grpc.NewServer()
	latisv1.RegisterLatisServiceServer(grpcServer, &server{})

	// Serve
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// server implements LatisServiceServer
type server struct {
	latisv1.UnimplementedLatisServiceServer
}

// Connect handles the bidirectional stream
func (s *server) Connect(stream latisv1.LatisService_ConnectServer) error {
	log.Println("new connection established")

	for {
		// Receive a message from cmdr
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("connection closed by client")
			return nil
		}
		if err != nil {
			log.Printf("error receiving: %v", err)
			return err
		}

		log.Printf("received message id=%s", req.Id)

		// Handle the message based on its type
		switch payload := req.Payload.(type) {
		case *latisv1.ConnectRequest_PromptSend:
			log.Printf("prompt: %s", payload.PromptSend.Content)

			// Echo the prompt back as response chunks
			content := fmt.Sprintf("Echo: %s", payload.PromptSend.Content)

			// Send a response chunk
			if err := stream.Send(&latisv1.ConnectResponse{
				Id: req.Id,
				Payload: &latisv1.ConnectResponse_ResponseChunk{
					ResponseChunk: &latisv1.ResponseChunk{
						RequestId: req.Id,
						Content:   content,
						Sequence:  0,
					},
				},
			}); err != nil {
				return err
			}

			// Send response complete
			if err := stream.Send(&latisv1.ConnectResponse{
				Id: req.Id,
				Payload: &latisv1.ConnectResponse_ResponseComplete{
					ResponseComplete: &latisv1.ResponseComplete{
						RequestId: req.Id,
					},
				},
			}); err != nil {
				return err
			}

		case *latisv1.ConnectRequest_Ping:
			log.Printf("ping received, timestamp=%d", payload.Ping.Timestamp)

			// Send pong
			if err := stream.Send(&latisv1.ConnectResponse{
				Id: req.Id,
				Payload: &latisv1.ConnectResponse_Pong{
					Pong: &latisv1.Pong{
						PingTimestamp: payload.Ping.Timestamp,
						PongTimestamp: time.Now().UnixNano(),
					},
				},
			}); err != nil {
				return err
			}

		default:
			log.Printf("unhandled message type: %T", payload)
		}
	}
}

// generateTLSConfig creates a self-signed certificate for development.
// In production, use proper certificates.
func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"latis"},
	}, nil
}

// Ensure we're using context (for future use)
var _ = context.Background
