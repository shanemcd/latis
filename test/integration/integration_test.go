package integration

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	latisv1 "github.com/shanemcd/latis/gen/go/latis/v1"
	"github.com/shanemcd/latis/pkg/pki"
	quictransport "github.com/shanemcd/latis/pkg/transport/quic"
)

// testServer implements LatisServiceServer for testing
type testServer struct {
	latisv1.UnimplementedLatisServiceServer
}

func (s *testServer) Connect(stream latisv1.LatisService_ConnectServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch payload := req.Payload.(type) {
		case *latisv1.ConnectRequest_PromptSend:
			content := fmt.Sprintf("Echo: %s", payload.PromptSend.Content)

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
		}
	}
}

// testEnv holds the test environment
type testEnv struct {
	ca         *pki.CA
	serverCert *pki.Cert
	clientCert *pki.Cert
	addr       string
	grpcServer *grpc.Server
	cleanup    func()
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Generate PKI
	ca, err := pki.GenerateCA()
	if err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}

	serverCert, err := pki.GenerateCert(ca, pki.UnitIdentity("test"), true, false)
	if err != nil {
		t.Fatalf("GenerateCert for server: %v", err)
	}

	clientCert, err := pki.GenerateCert(ca, pki.CmdrIdentity(), false, true)
	if err != nil {
		t.Fatalf("GenerateCert for client: %v", err)
	}

	// Create server TLS config
	serverTLS, err := pki.ServerTLSConfig(serverCert, ca)
	if err != nil {
		t.Fatalf("ServerTLSConfig: %v", err)
	}

	// Start QUIC listener on random port
	listener, err := quictransport.Listen("127.0.0.1:0", serverTLS, nil)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}

	// Create and start gRPC server
	grpcServer := grpc.NewServer()
	latisv1.RegisterLatisServiceServer(grpcServer, &testServer{})

	go func() {
		grpcServer.Serve(listener)
	}()

	return &testEnv{
		ca:         ca,
		serverCert: serverCert,
		clientCert: clientCert,
		addr:       listener.Addr().String(),
		grpcServer: grpcServer,
		cleanup: func() {
			grpcServer.Stop()
			listener.Close()
		},
	}
}

func (e *testEnv) connect(t *testing.T) (latisv1.LatisServiceClient, func()) {
	t.Helper()

	clientTLS, err := pki.ClientTLSConfig(e.clientCert, e.ca, "localhost")
	if err != nil {
		t.Fatalf("ClientTLSConfig: %v", err)
	}

	dialer := quictransport.NewDialer(clientTLS, nil)

	conn, err := grpc.NewClient(
		e.addr,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	client := latisv1.NewLatisServiceClient(conn)
	return client, func() { conn.Close() }
}

func TestPing(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	client, cleanup := env.connect(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Send ping
	sendTime := time.Now().UnixNano()
	if err := stream.Send(&latisv1.ConnectRequest{
		Id: uuid.New().String(),
		Payload: &latisv1.ConnectRequest_Ping{
			Ping: &latisv1.Ping{
				Timestamp: sendTime,
			},
		},
	}); err != nil {
		t.Fatalf("Send ping: %v", err)
	}

	// Receive pong
	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv: %v", err)
	}

	pong, ok := resp.Payload.(*latisv1.ConnectResponse_Pong)
	if !ok {
		t.Fatalf("Expected Pong, got %T", resp.Payload)
	}

	if pong.Pong.PingTimestamp != sendTime {
		t.Errorf("PingTimestamp = %d, want %d", pong.Pong.PingTimestamp, sendTime)
	}

	latency := time.Duration(time.Now().UnixNano() - sendTime)
	t.Logf("Round-trip latency: %v", latency)

	if latency > 1*time.Second {
		t.Errorf("Latency too high: %v", latency)
	}
}

func TestPrompt(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	client, cleanup := env.connect(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Send prompt
	promptContent := "Hello, integration test!"
	msgID := uuid.New().String()

	if err := stream.Send(&latisv1.ConnectRequest{
		Id: msgID,
		Payload: &latisv1.ConnectRequest_PromptSend{
			PromptSend: &latisv1.PromptSend{
				Content: promptContent,
			},
		},
	}); err != nil {
		t.Fatalf("Send prompt: %v", err)
	}

	// Receive chunk
	resp1, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv chunk: %v", err)
	}

	chunk, ok := resp1.Payload.(*latisv1.ConnectResponse_ResponseChunk)
	if !ok {
		t.Fatalf("Expected ResponseChunk, got %T", resp1.Payload)
	}

	expectedContent := "Echo: " + promptContent
	if chunk.ResponseChunk.Content != expectedContent {
		t.Errorf("Content = %q, want %q", chunk.ResponseChunk.Content, expectedContent)
	}

	// Receive complete
	resp2, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv complete: %v", err)
	}

	complete, ok := resp2.Payload.(*latisv1.ConnectResponse_ResponseComplete)
	if !ok {
		t.Fatalf("Expected ResponseComplete, got %T", resp2.Payload)
	}

	if complete.ResponseComplete.RequestId != msgID {
		t.Errorf("RequestId = %q, want %q", complete.ResponseComplete.RequestId, msgID)
	}
}

func TestMultipleMessages(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	client, cleanup := env.connect(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Send multiple pings
	numPings := 5
	for i := 0; i < numPings; i++ {
		if err := stream.Send(&latisv1.ConnectRequest{
			Id: uuid.New().String(),
			Payload: &latisv1.ConnectRequest_Ping{
				Ping: &latisv1.Ping{
					Timestamp: time.Now().UnixNano(),
				},
			},
		}); err != nil {
			t.Fatalf("Send ping %d: %v", i, err)
		}
	}

	// Receive all pongs
	for i := 0; i < numPings; i++ {
		resp, err := stream.Recv()
		if err != nil {
			t.Fatalf("Recv pong %d: %v", i, err)
		}

		if _, ok := resp.Payload.(*latisv1.ConnectResponse_Pong); !ok {
			t.Fatalf("Expected Pong %d, got %T", i, resp.Payload)
		}
	}

	t.Logf("Successfully exchanged %d ping/pong messages", numPings)
}

func TestConnectionRequiresMTLS(t *testing.T) {
	env := setupTestEnv(t)
	defer env.cleanup()

	// Try to connect without proper client cert
	wrongCA, err := pki.GenerateCA()
	if err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}

	wrongClientCert, err := pki.GenerateCert(wrongCA, pki.CmdrIdentity(), false, true)
	if err != nil {
		t.Fatalf("GenerateCert: %v", err)
	}

	// Use wrong CA's cert as client cert (signed by different CA)
	clientTLS, err := pki.ClientTLSConfig(wrongClientCert, wrongCA, "localhost")
	if err != nil {
		t.Fatalf("ClientTLSConfig: %v", err)
	}

	dialer := quictransport.NewDialer(clientTLS, nil)

	conn, err := grpc.NewClient(
		env.addr,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		// Connection creation might fail
		t.Logf("NewClient failed (expected): %v", err)
		return
	}
	defer conn.Close()

	client := latisv1.NewLatisServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Connect(ctx)
	if err == nil {
		t.Error("Expected connection to fail with wrong CA, but it succeeded")
	} else {
		t.Logf("Connection correctly rejected: %v", err)
	}
}
