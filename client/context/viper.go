package context

import (
	"fmt"

	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// NewCoreContextFromViper - return a new context with parameters from the command line
func NewCoreContextFromViper() CoreContext {
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
	return CoreContext{
		ChainID:         chainID,
		Height:          viper.GetInt64(client.FlagHeight),
		Gas:             viper.GetInt64(client.FlagGas),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		FromAddressName: viper.GetString(client.FlagName),
		NodeURI:         nodeURI,
		AccountNumber:   viper.GetInt64(client.FlagAccountNumber),
		Sequence:        viper.GetInt64(client.FlagSequence),
		Client:          rpc,
		Decoder:         nil,
		AccountStore:    "acc",
	}
}

// read chain ID from genesis file, if present
func defaultChainID() (string, error) {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return "", err
	}
	doc, err := tmtypes.GenesisDocFromFile(cfg.GenesisFile())
	if err != nil {
		return "", err
	}
	return doc.ChainID, nil
}

// EnsureSequence - automatically set sequence number if none provided
func EnsureAccountNumber(ctx CoreContext) (CoreContext, error) {
	// Should be viper.IsSet, but this does not work - https://github.com/spf13/viper/pull/331
	if viper.GetInt64(client.FlagAccountNumber) != 0 {
		return ctx, nil
	}
	from, err := ctx.GetFromAddress()
	if err != nil {
		return ctx, err
	}
	accnum, err := ctx.GetAccountNumber(from)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("Defaulting to account number: %d\n", accnum)
	ctx = ctx.WithAccountNumber(accnum)
	return ctx, nil
}

// EnsureSequence - automatically set sequence number if none provided
func EnsureSequence(ctx CoreContext) (CoreContext, error) {
	// Should be viper.IsSet, but this does not work - https://github.com/spf13/viper/pull/331
	if viper.GetInt64(client.FlagSequence) != 0 {
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
