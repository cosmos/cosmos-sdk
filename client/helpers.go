package client

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	cmn "github.com/tendermint/tmlibs/common"
)

// GetNode prepares a simple rpc.Client from the flags
func GetNode() (rpcclient.Client, error) {
	uri := viper.GetString(FlagNode)
	if uri == "" {
		return nil, errors.New("Must define node using --node")
	}
	return rpcclient.NewHTTP(uri, "/websocket"), nil
}

// Broadcast the transaction bytes to Tendermint
func BroadcastTx(tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {

	node, err := GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return res, err
	}

	if res.CheckTx.Code != uint32(0) {
		return res, errors.Errorf("CheckTx failed: (%d) %s",
			res.CheckTx.Code,
			res.CheckTx.Log)
	}
	if res.DeliverTx.Code != uint32(0) {
		return res, errors.Errorf("DeliverTx failed: (%d) %s",
			res.DeliverTx.Code,
			res.DeliverTx.Log)
	}
	return res, err
}

// Query from Tendermint with the provided key and storename
func Query(key cmn.HexBytes, storeName string) (res []byte, err error) {

	path := fmt.Sprintf("/%s/key", storeName)
	node, err := GetNode()
	if err != nil {
		return res, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height:  viper.GetInt64(FlagHeight),
		Trusted: viper.GetBool(FlagTrustNode),
	}
	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}
	resp := result.Response
	if resp.Code != uint32(0) {
		return res, errors.Errorf("Query failed: (%d) %s", resp.Code, resp.Log)
	}
	return resp.Value, nil
}
