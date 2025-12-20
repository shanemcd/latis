// Package provisioner defines the interface for unit lifecycle management.
// Provisioners create, start, stop, and destroy units (processes, containers, VMs, etc.).
package provisioner

import "context"

// Provisioner manages the lifecycle of a unit.
type Provisioner interface {
	// Create provisions a new unit (but does not start it).
	Create(ctx context.Context) error

	// Start starts a provisioned unit.
	Start(ctx context.Context) error

	// Stop stops a running unit.
	Stop(ctx context.Context) error

	// Destroy removes a unit and cleans up resources.
	Destroy(ctx context.Context) error
}
