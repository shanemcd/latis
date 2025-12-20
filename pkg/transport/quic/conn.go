package quic

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// Conn wraps a quic.Stream to implement net.Conn.
// This allows gRPC to use QUIC streams as if they were TCP connections.
type Conn struct {
	stream *quic.Stream
	qconn  *quic.Conn
	local  net.Addr
	remote net.Addr
}

// Read reads data from the QUIC stream.
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write writes data to the QUIC stream.
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close closes the QUIC stream and connection.
func (c *Conn) Close() error {
	// Close the stream first
	if err := c.stream.Close(); err != nil {
		return err
	}
	// Then close the connection
	return c.qconn.CloseWithError(0, "connection closed")
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.local
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.remote
}

// SetDeadline sets the read and write deadlines.
func (c *Conn) SetDeadline(t time.Time) error {
	if err := c.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return c.stream.SetWriteDeadline(t)
}

// SetReadDeadline sets the read deadline.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// Ensure Conn implements net.Conn
var _ net.Conn = (*Conn)(nil)
