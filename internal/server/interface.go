package server

import (
	"context"
	"net"
)

// Server Interface
type Server interface {
	Start(ctx context.Context) error
	Stop()
	GetAddr() net.Addr
}
