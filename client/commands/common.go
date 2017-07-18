/*
Package commands contains any general setup/helpers valid for all subcommands
*/
package commands

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tmlibs/cli"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/tendermint/light-client/certifiers"
	"github.com/tendermint/light-client/certifiers/client"
	"github.com/tendermint/light-client/certifiers/files"
)

var (
	trustedProv certifiers.Provider
	sourceProv  certifiers.Provider
)

const (
	ChainFlag = "chain-id"
	NodeFlag  = "node"
)

func AddBasicFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(ChainFlag, "", "Chain ID of tendermint node")
	cmd.PersistentFlags().String(NodeFlag, "", "<host>:<port> to tendermint rpc interface for this chain")
}

func GetChainID() string {
	return viper.GetString(ChainFlag)
}

func GetNode() rpcclient.Client {
	return rpcclient.NewHTTP(viper.GetString(NodeFlag), "/websocket")
}

func GetProviders() (trusted certifiers.Provider, source certifiers.Provider) {
	if trustedProv == nil || sourceProv == nil {
		// initialize provider with files stored in homedir
		rootDir := viper.GetString(cli.HomeFlag)
		trustedProv = certifiers.NewCacheProvider(
			certifiers.NewMemStoreProvider(),
			files.NewProvider(rootDir),
		)
		node := viper.GetString(NodeFlag)
		sourceProv = client.NewHTTP(node)
	}
	return trustedProv, sourceProv
}

func GetCertifier() (*certifiers.InquiringCertifier, error) {
	// load up the latest store....
	trust, source := GetProviders()

	// this gets the most recent verified seed
	seed, err := certifiers.LatestSeed(trust)
	if certifiers.IsSeedNotFoundErr(err) {
		return nil, errors.New("Please run init first to establish a root of trust")
	}
	if err != nil {
		return nil, err
	}
	cert := certifiers.NewInquiring(
		viper.GetString(ChainFlag), seed.Validators, trust, source)
	return cert, nil
}
