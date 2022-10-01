package main

import (
	"fmt"

	"github.com/hashicorp/go-plugin"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/grpc_abci_v1"
)

// ABCIListener is the implementation of the baseapp.ABCIListener interface
type ABCIListener struct {
	blockHeight    int64
	txIdx          int64
	storeKVPairIdx int64
}

func (m *ABCIListener) ListenBeginBlock(blockHeight int64, req []byte, res []byte) error {
	fmt.Printf("begin-block: block_height=%d req=%b res=%s", blockHeight, req, res)
	return nil
}

func (m *ABCIListener) ListenEndBlock(blockHeight int64, req []byte, res []byte) error {
	fmt.Printf("end-block: block_height=%d req=%b res=%s", blockHeight, req, res)
	return nil
}

func (m *ABCIListener) ListenDeliverTx(blockHeight int64, req []byte, res []byte) error {
	m.updateTxIdx(blockHeight)
	fmt.Printf("deliver-tx: block_height=%d idx: %d, req=%b res=%s", blockHeight, m.txIdx, req, res)
	return nil
}

func (m *ABCIListener) ListenStoreKVPair(blockHeight int64, data []byte) error {
	m.updateStoreKVPairIdx(blockHeight)
	fmt.Printf("store-kv-pair: block_height=%d idx: %d, req=%b", blockHeight, m.storeKVPairIdx, data)
	return nil
}

func (m *ABCIListener) updateTxIdx(currBlockHeight int64) {
	if m.blockHeight < currBlockHeight {
		m.txIdx = 0
	} else {
		m.txIdx++
	}
}

func (m *ABCIListener) updateStoreKVPairIdx(currBlockHeight int64) {
	if m.blockHeight < currBlockHeight {
		m.storeKVPairIdx = 0
	} else {
		m.storeKVPairIdx++
	}
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: grpc_abci_v1.Handshake,
		Plugins: map[string]plugin.Plugin{
			"grpc_plugin_v1": &grpc_abci_v1.ListenerGRPCPlugin{Impl: &ABCIListener{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
