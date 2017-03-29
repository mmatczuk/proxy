package proxy

import "context"

// RemoteClient provides ability to call the legacy system, implementations must
// be thread safe.
type RemoteClient interface {
	Update(ctx context.Context, addr, info string) error
}
