// Package streaming provides shared data structures and interfaces for communication
// between the host application and plugins in a streaming context.
package streaming

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Listener defines the interface for a streaming service that hooks into
// the ABCI message processing of the BaseApp. Implementations should handle
// errors internally and return nil if they don't want to affect consensus.
type Listener interface {
	// ListenDeliverBlock updates the streaming service with the latest Delivered Block messages.
	ListenDeliverBlock(context.Context, ListenDeliverBlockRequest) error

	// ListenStateChanges updates the streaming service with the latest Commit messages and state changes.
	ListenStateChanges(ctx context.Context, changeSet []*StoreKVPair) error
}

// Handshake defines the handshake configuration shared by the streaming service and host.
// It serves as a UX feature to prevent execution of incompatible or unintended plugins.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ABCI_LISTENER_PLUGIN",
	MagicCookieValue: "ef78114d-7bdf-411c-868f-347c99a78345",
}

var _ plugin.GRPCPlugin = (*ListenerGRPCPlugin)(nil)

// ListenerGRPCPlugin is the implementation of plugin.GRPCPlugin, so we can serve/consume this.
type ListenerGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Listener
}

// GRPCServer registers the ListenerService server implementation.
func (p *ListenerGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterListenerServiceServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient creates a new ListenerService client.
func (p *ListenerGRPCPlugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &GRPCClient{client: NewListenerServiceClient(c)}, nil
}
