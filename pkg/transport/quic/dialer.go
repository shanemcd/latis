package quic

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
)

// Dial establishes a QUIC connection and returns a net.Conn.
// The returned Conn wraps a QUIC stream suitable for gRPC.
func Dial(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (net.Conn, error) {
	// Establish QUIC connection
	qconn, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}

	// Open a stream for gRPC communication
	stream, err := qconn.OpenStreamSync(ctx)
	if err != nil {
		qconn.CloseWithError(1, "failed to open stream")
		return nil, err
	}

	return &Conn{
		stream: stream,
		qconn:  qconn,
		local:  qconn.LocalAddr(),
		remote: qconn.RemoteAddr(),
	}, nil
}

// Dialer is a function type that matches grpc.WithContextDialer expectations.
// Use this with grpc.WithContextDialer(quictransport.NewDialer(tlsConfig, quicConfig))
type Dialer func(ctx context.Context, addr string) (net.Conn, error)

// NewDialer returns a Dialer function configured with the given TLS and QUIC configs.
func NewDialer(tlsConfig *tls.Config, quicConfig *quic.Config) Dialer {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return Dial(ctx, addr, tlsConfig, quicConfig)
	}
}
