// Package dialer defines the interface for outbound connections from cmdr to units.
package dialer

import (
	"context"

	"github.com/shanemcd/latis/pkg/connector"
)

// Dialer establishes outbound connections to units.
type Dialer interface {
	// Dial connects to a unit at the given address.
	Dial(ctx context.Context, address string) (connector.Connector, error)
}
