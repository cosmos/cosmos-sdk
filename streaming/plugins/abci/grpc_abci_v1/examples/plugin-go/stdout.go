package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/grpc_abci_v1"
)

// ABCIListenerPlugin is the implementation of the baseapp.ABCIListener interface
type ABCIListenerPlugin struct{}

func (m *ABCIListenerPlugin) Listen(ctx context.Context, blockHeight int64, eventType string, data []byte) error {
	fmt.Printf("event-type:%s block_height=%d data=%b", eventType, blockHeight, data)
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: grpc_abci_v1.Handshake,
		Plugins: map[string]plugin.Plugin{
			"grpc_plugin_v1": &grpc_abci_v1.ABCIListenerGRPCPlugin{Impl: &ABCIListenerPlugin{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
