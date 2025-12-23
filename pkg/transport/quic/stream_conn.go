package quic

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// StreamConn wraps a QUIC stream to implement net.Conn.
// Unlike Conn, closing a StreamConn only closes the stream, not the connection.
type StreamConn struct {
	stream     *quic.Stream
	qconn      *quic.Conn
	local      net.Addr
	remote     net.Addr
	streamType StreamType
}

// Read reads data from the QUIC stream.
func (c *StreamConn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

// Write writes data to the QUIC stream.
func (c *StreamConn) Write(b []byte) (int, error) {
	return c.stream.Write(b)
}

// Close closes the QUIC stream only (not the connection).
func (c *StreamConn) Close() error {
	return c.stream.Close()
}

// LocalAddr returns the local network address.
func (c *StreamConn) LocalAddr() net.Addr {
	return c.local
}

// RemoteAddr returns the remote network address.
func (c *StreamConn) RemoteAddr() net.Addr {
	return c.remote
}

// SetDeadline sets the read and write deadlines.
func (c *StreamConn) SetDeadline(t time.Time) error {
	if err := c.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return c.stream.SetWriteDeadline(t)
}

// SetReadDeadline sets the read deadline.
func (c *StreamConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline.
func (c *StreamConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// StreamType returns the type of this stream.
func (c *StreamConn) StreamType() StreamType {
	return c.streamType
}

// StreamID returns the QUIC stream ID.
func (c *StreamConn) StreamID() quic.StreamID {
	return c.stream.StreamID()
}

// Ensure StreamConn implements net.Conn
var _ net.Conn = (*StreamConn)(nil)
