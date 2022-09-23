// Package abci contains shared data between the host and plugins.
package abci

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/proto"
)

// Handshake is a common handshake that is shared by streaming and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "ABCI_LISTENER_PLUGIN",
	MagicCookieValue: "hello",
}

// Listener is the interface that we're exposing as a streaming.
type Listener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(blockHeight int64, req []byte, res []byte) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(blockHeight int64, req []byte, res []byte) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(blockHeight int64, req []byte, res []byte) error
	// ListenStoreKVPair updates the steaming service with the latest StoreKVPair messages
	ListenStoreKVPair(blockHeight int64, data []byte) error
}

// ListenerPlugin is the implementation of streaming.Plugin so we can serve/consume this.
type ListenerPlugin struct {
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Listener
}

// ListenerGRPCPlugin is the implementation of plugin.GRPCPlugin, so we can serve/consume this.
type ListenerGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Listener
}

func (p *ListenerGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterABCIListenerServiceServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *ListenerGRPCPlugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &GRPCClient{client: proto.NewABCIListenerServiceClient(c)}, nil
}
