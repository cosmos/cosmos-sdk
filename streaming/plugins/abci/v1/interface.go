// Package v1 contains shared data between the host and plugins.
package v1

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// Handshake is a common handshake that is shared by streaming and host.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "ABCI_LISTENER_PLUGIN",
	MagicCookieValue: "ef78114d-7bdf-411c-868f-347c99a78345",
}

var _ plugin.GRPCPlugin = (*ABCIListenerGRPCPlugin)(nil)

// ABCIListenerGRPCPlugin is the implementation of plugin.GRPCPlugin, so we can serve/consume this.
type ABCIListenerGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl baseapp.ABCIListener
}

func (p *ABCIListenerGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterABCIListenerServiceServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *ABCIListenerGRPCPlugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &GRPCClient{client: NewABCIListenerServiceClient(c)}, nil
}
