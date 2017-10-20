/*
Package commands contains any general setup/helpers valid for all subcommands
*/
package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/modules/auth"
)

var (
	trustedProv certifiers.Provider
	sourceProv  certifiers.Provider
)

const (
	ChainFlag = "chain-id"
	NodeFlag  = "node"
)

// AddBasicFlags adds --node and --chain-id, which we need for everything
func AddBasicFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(ChainFlag, "", "Chain ID of tendermint node")
	cmd.PersistentFlags().String(NodeFlag, "", "<host>:<port> to tendermint rpc interface for this chain")
}

// GetChainID reads ChainID from the flags
func GetChainID() string {
	return viper.GetString(ChainFlag)
}

// GetNode prepares a simple rpc.Client from the flags
func GetNode() rpcclient.Client {
	return client.GetNode(viper.GetString(NodeFlag))
}

// GetSourceProvider returns a provider pointing to an rpc handler
func GetSourceProvider() certifiers.Provider {
	if sourceProv == nil {
		node := viper.GetString(NodeFlag)
		sourceProv = client.GetRPCProvider(node)
	}
	return sourceProv
}

// GetTrustedProvider returns a reference to a local store with cache
func GetTrustedProvider() certifiers.Provider {
	if trustedProv == nil {
		rootDir := viper.GetString(cli.HomeFlag)
		trustedProv = client.GetLocalProvider(rootDir)
	}
	return trustedProv
}

// GetProviders creates a trusted (local) seed provider and a remote
// provider based on configuration.
func GetProviders() (trusted certifiers.Provider, source certifiers.Provider) {
	return GetTrustedProvider(), GetSourceProvider()
}

// GetCertifier constructs a dynamic certifier from the config info
func GetCertifier() (*certifiers.Inquiring, error) {
	// load up the latest store....
	trust := GetTrustedProvider()
	source := GetSourceProvider()
	chainID := GetChainID()
	return client.GetCertifier(chainID, trust, source)
}

// ParseActor parses an address of form:
// [<chain>:][<app>:]<hex address>
// into a sdk.Actor.
// If app is not specified or "", then assume auth.NameSigs
func ParseActor(input string) (res sdk.Actor, err error) {
	chain, app := "", auth.NameSigs
	input = strings.TrimSpace(input)
	spl := strings.SplitN(input, ":", 3)

	if len(spl) == 3 {
		chain = spl[0]
		spl = spl[1:]
	}
	if len(spl) == 2 {
		if spl[0] != "" {
			app = spl[0]
		}
		spl = spl[1:]
	}

	addr, err := hex.DecodeString(cmn.StripHex(spl[0]))
	if err != nil {
		return res, errors.Errorf("Address is invalid hex: %v\n", err)
	}
	res = sdk.Actor{
		ChainID: chain,
		App:     app,
		Address: addr,
	}
	return
}

// ParseActors takes a comma-separated list of actors and parses them into
// a slice
func ParseActors(key string) (signers []sdk.Actor, err error) {
	var act sdk.Actor
	for _, k := range strings.Split(key, ",") {
		act, err = ParseActor(k)
		if err != nil {
			return
		}
		signers = append(signers, act)
	}
	return
}

// GetOneArg makes sure there is exactly one positional argument
func GetOneArg(args []string, argname string) (string, error) {
	if len(args) == 0 {
		return "", errors.Errorf("Missing required argument [%s]", argname)
	}
	if len(args) > 1 {
		return "", errors.Errorf("Only accepts one argument [%s]", argname)
	}
	return args[0], nil
}

// ParseHexFlag takes a flag name and parses the viper contents as hex
func ParseHexFlag(flag string) ([]byte, error) {
	arg := viper.GetString(flag)
	if arg == "" {
		return nil, errors.Errorf("No such flag: %s", flag)
	}
	value, err := hex.DecodeString(cmn.StripHex(arg))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Cannot parse %s", flag))
	}
	return value, nil

}
