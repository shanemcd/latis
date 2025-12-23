package quic

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

// MuxConn wraps a QUIC connection and provides multiplexed stream access.
// It routes incoming streams by type and provides methods to open typed streams.
type MuxConn struct {
	qconn  *quic.Conn
	local  net.Addr
	remote net.Addr

	mu       sync.Mutex
	closed   bool
	closeErr error
}

// NewMuxConn wraps a QUIC connection for multiplexed stream handling.
func NewMuxConn(qconn *quic.Conn) *MuxConn {
	return &MuxConn{
		qconn:  qconn,
		local:  qconn.LocalAddr(),
		remote: qconn.RemoteAddr(),
	}
}

// OpenStream opens a new stream of the given type.
// The stream type byte is written as the first byte.
func (c *MuxConn) OpenStream(ctx context.Context, streamType StreamType) (net.Conn, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("connection closed: %w", c.closeErr)
	}
	c.mu.Unlock()

	stream, err := c.qconn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	// Write stream type as first byte
	if _, err := stream.Write([]byte{byte(streamType)}); err != nil {
		stream.Close()
		return nil, fmt.Errorf("write stream type: %w", err)
	}

	return &StreamConn{
		stream:     stream,
		qconn:      c.qconn,
		local:      c.local,
		remote:     c.remote,
		streamType: streamType,
	}, nil
}

// AcceptStream accepts an incoming stream and reads its type.
// Returns the stream wrapped as net.Conn and its type.
func (c *MuxConn) AcceptStream(ctx context.Context) (net.Conn, StreamType, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, 0, fmt.Errorf("connection closed: %w", c.closeErr)
	}
	c.mu.Unlock()

	stream, err := c.qconn.AcceptStream(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("accept stream: %w", err)
	}

	// Read stream type from first byte
	typeBuf := make([]byte, 1)
	if _, err := io.ReadFull(stream, typeBuf); err != nil {
		stream.Close()
		return nil, 0, fmt.Errorf("read stream type: %w", err)
	}
	streamType := StreamType(typeBuf[0])

	return &StreamConn{
		stream:     stream,
		qconn:      c.qconn,
		local:      c.local,
		remote:     c.remote,
		streamType: streamType,
	}, streamType, nil
}

// LocalAddr returns the local network address.
func (c *MuxConn) LocalAddr() net.Addr {
	return c.local
}

// RemoteAddr returns the remote network address.
func (c *MuxConn) RemoteAddr() net.Addr {
	return c.remote
}

// Close closes the underlying QUIC connection.
func (c *MuxConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return c.closeErr
	}
	c.closed = true
	c.closeErr = c.qconn.CloseWithError(0, "connection closed")
	return c.closeErr
}

// Context returns the connection's context, which is canceled when the connection is closed.
func (c *MuxConn) Context() context.Context {
	return c.qconn.Context()
}
