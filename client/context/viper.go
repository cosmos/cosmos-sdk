package context

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/core"
)

// NewCoreContextFromViper - return a new context with parameters from the command line
func NewCoreContextFromViper() core.CoreContext {
	nodeURI := viper.GetString(client.FlagNode)
	var rpc rpcclient.Client
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}
	chainID := viper.GetString(client.FlagChainID)
	// if chain ID is not specified manually, read default chain ID
	if chainID == "" {
		def, err := defaultChainID()
		if err != nil {
			chainID = def
		}
	}
	return core.CoreContext{
		ChainID:         chainID,
		Height:          viper.GetInt64(client.FlagHeight),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		FromAddressName: viper.GetString(client.FlagName),
		NodeURI:         nodeURI,
		Sequence:        viper.GetInt64(client.FlagSequence),
		Client:          rpc,
		Decoder:         nil,
		AccountStore:    "main",
	}
}

// read chain ID from genesis file, if present
func defaultChainID() (string, error) {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return "", err
	}
	genesisFile := cfg.GenesisFile()
	bz, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return "", err
	}
	var doc tmtypes.GenesisDoc
	err = json.Unmarshal(bz, &doc)
	if err != nil {
		return "", err
	}
	return doc.ChainID, nil
}

// EnsureSequence - automatically set sequence number if none provided
func EnsureSequence(ctx core.CoreContext) (core.CoreContext, error) {
	if viper.IsSet(client.FlagSequence) {
		return ctx, nil
	}
	from, err := ctx.GetFromAddress()
	if err != nil {
		return ctx, err
	}
	seq, err := ctx.NextSequence(from)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("Defaulting to next sequence number: %d\n", seq)
	ctx = ctx.WithSequence(seq)
	return ctx, nil
}
