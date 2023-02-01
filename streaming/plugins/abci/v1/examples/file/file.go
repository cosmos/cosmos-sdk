package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-plugin"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/v1"
	abci "github.com/tendermint/tendermint/abci/types"
)

// FilePlugin is the implementation of the ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type FilePlugin struct {
	BlockHeight int64
	fileDict    map[string]*os.File
}

func (a *FilePlugin) writeToFile(file string, data []byte) error {
	if f, ok := a.fileDict[file]; ok {
		_, err := f.Write(data)
		return err
	}
	a.fileDict = make(map[string]*os.File, 0)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s.txt", home, file)
	f, err := os.OpenFile(filepath.Clean(filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	a.fileDict[file] = f

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

func (a *FilePlugin) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	a.BlockHeight = req.Header.Height
	d1 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, req))
	d2 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, res))
	if err := a.writeToFile("begin-block-req", d1); err != nil {
		return err
	}
	if err := a.writeToFile("begin-block-res", d2); err != nil {
		return err
	}
	return nil
}

func (a *FilePlugin) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	d1 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, req))
	d2 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, req))
	if err := a.writeToFile("end-block-req", d1); err != nil {
		return err
	}
	if err := a.writeToFile("end-block-res", d2); err != nil {
		return err
	}
	return nil
}

func (a *FilePlugin) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	d1 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, req))
	d2 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, res))
	if err := a.writeToFile("deliver-tx-req", d1); err != nil {
		return err
	}
	if err := a.writeToFile("deliver-tx-res", d2); err != nil {
		return err
	}
	return nil
}

func (a *FilePlugin) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	fmt.Printf("listen-commit: block_height=%d data=%v", res.RetainHeight, changeSet)
	d1 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, res))
	d2 := []byte(fmt.Sprintf("%d:::%v\n", a.BlockHeight, changeSet))
	if err := a.writeToFile("commit-res", d1); err != nil {
		return err
	}
	if err := a.writeToFile("state-change", d2); err != nil {
		return err
	}
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: v1.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci_v1": &v1.ABCIListenerGRPCPlugin{Impl: &FilePlugin{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
