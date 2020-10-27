package server

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/tendermint/cosmos-rosetta-gateway/service"
)

const (
	flagBlockchain    = "blockchain"
	flagNetwork       = "network"
	flagTendermintRpc = "tendermint-rpc"
	flagAppRpc        = "app-rpc"
	flagOfflineMode   = "offline"
)

// RosettaCommand will start the application Rosetta API service as a blocking process.
func RosettaCommand(cdc codec.BinaryMarshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use: "rosetta",
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := getRosettaOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			s, err := service.New(
				service.Options{Port: 8080},
				rosetta.NewNetwork(cdc, options),
			)
			if err != nil {
				panic(err)
			}

			err = s.Start()
			if err != nil {
				panic(err)
			}

			return nil
		},
	}

	cmd.Flags().String(flagBlockchain, "blockchain", "Application's name (e.g. Cosmos Hub)")
	cmd.Flags().String(flagNetwork, "network", "Network's identifier (e.g. cosmos-hub-3, testnet-1, etc)")
	cmd.Flags().String(flagAppRpc, "localhost:1317", "Application's RPC endpoint.")
	cmd.Flags().String(flagTendermintRpc, "localhost:26657", "Tendermint's RPC endpoint.")
	cmd.Flags().Bool(flagOfflineMode, false, "Flag that forces the rosetta service to run in offline mode, some endpoints won't work.")

	return cmd
}

func getRosettaOptionsFromFlags(flags *flag.FlagSet) (rosetta.Options, error) {
	blockchain, err := flags.GetString(flagBlockchain)
	if err != nil {
		return rosetta.Options{}, fmt.Errorf("invalid blockchain value: %w", err)
	}

	network, err := flags.GetString(flagNetwork)
	if err != nil {
		return rosetta.Options{}, fmt.Errorf("invalid network value: %w", err)
	}

	appRpc, err := flags.GetString(flagAppRpc)
	if err != nil {
		return rosetta.Options{}, fmt.Errorf("invalid app rpc value: %w", err)
	}

	tendermintRpc, err := flags.GetString(flagTendermintRpc)
	if err != nil {
		return rosetta.Options{}, fmt.Errorf("invalid tendermint rpc value: %w", err)
	}

	offline, err := flags.GetBool(flagOfflineMode)
	if err != nil {
		return rosetta.Options{}, fmt.Errorf("invalid offline value: %w", err)
	}

	return rosetta.Options{
		AppEndpoint:        appRpc,
		TendermintEndpoint: tendermintRpc,
		Blockchain:         blockchain,
		Network:            network,
		OfflineMode:        offline,
	}, nil
}
