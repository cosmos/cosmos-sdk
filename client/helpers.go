package client

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// GetNode prepares a simple rpc.Client from the flags
func GetNode() (rpcclient.Client, error) {
	uri := viper.GetString(FlagNode)
	if uri == "" {
		return nil, errors.New("Must define node using --node")
	}
	return rpcclient.NewHTTP(uri, "/websocket"), nil
}
