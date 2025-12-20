// Package quic provides QUIC transport adapters for gRPC.
//
// This package wraps quic-go types to implement standard net.Listener and
// net.Conn interfaces, allowing gRPC to run over QUIC without modification.
//
// # Design
//
// QUIC connections contain multiple streams. We use one stream per gRPC
// connection. The stream provides the bidirectional byte flow that gRPC
// expects from a net.Conn.
//
// # Usage
//
// Server:
//
//	listener, err := quictransport.Listen(addr, tlsConfig, quicConfig)
//	grpcServer.Serve(listener)
//
// Client:
//
//	conn, err := quictransport.Dial(ctx, addr, tlsConfig, quicConfig)
//	grpcConn, err := grpc.NewClient("passthrough:///"+addr, grpc.WithTransportCredentials(...), ...)
package quic

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
)

// Listener wraps a quic.Listener to implement net.Listener.
// Each Accept() returns a Conn wrapping a new QUIC stream.
type Listener struct {
	ql        *quic.Listener
	tlsConfig *tls.Config
}

// Listen creates a new QUIC listener on the given address.
func Listen(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (*Listener, error) {
	ql, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}
	return &Listener{ql: ql, tlsConfig: tlsConfig}, nil
}

// Accept waits for and returns the next connection.
// It accepts a QUIC connection, then accepts a stream from it.
// The stream is wrapped as a net.Conn.
func (l *Listener) Accept() (net.Conn, error) {
	ctx := context.Background()

	// Accept a new QUIC connection
	qconn, err := l.ql.Accept(ctx)
	if err != nil {
		return nil, err
	}

	// Accept the first stream from this connection
	// The client is expected to open a stream immediately after connecting
	stream, err := qconn.AcceptStream(ctx)
	if err != nil {
		qconn.CloseWithError(1, "failed to accept stream")
		return nil, err
	}

	return &Conn{
		stream: stream,
		qconn:  qconn,
		local:  qconn.LocalAddr(),
		remote: qconn.RemoteAddr(),
	}, nil
}

// Close closes the listener.
func (l *Listener) Close() error {
	return l.ql.Close()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.ql.Addr()
}

// Ensure Listener implements net.Listener
var _ net.Listener = (*Listener)(nil)
