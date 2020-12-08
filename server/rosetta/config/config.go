package config

import (
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos"
	"github.com/cosmos/cosmos-sdk/server/rosetta/services"
	"github.com/cosmos/cosmos-sdk/server/rosetta/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/tendermint/cosmos-rosetta-gateway/service"
)

const (
	flagBlockchain    = "blockchain"
	flagNetwork       = "network"
	flagTendermintRPC = "tendermint-rpc"
	flagAppRPC        = "app-rpc"
	flagOfflineMode   = "offline"
	flagListenAddr    = "listen-addr"
)

type Options struct {
	AppEndpoint        string
	TendermintEndpoint string
	Blockchain         string
	Network            string
	OfflineMode        bool
}

// RosettaCommand will start the application Rosetta API service as a blocking process.
func RosettaCommand(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "rosetta",
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := getRosettaOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			listenAddr, err := cmd.Flags().GetString(flagListenAddr)
			if err != nil {
				return err
			}

			var client rosetta.CosmosClient
			err = util.TryFunction(5, 5*time.Second, func() (err error) {
				client, err = cosmos.NewDataClient(options.TendermintEndpoint, options.AppEndpoint, cdc)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}

			svc, err := services.NewOnline(client, &types.NetworkIdentifier{
				Blockchain:           options.Blockchain,
				Network:              options.Network,
				SubNetworkIdentifier: nil,
			})
			if err != nil {
				return err
			}

			s, err := service.New(
				service.Options{ListenAddress: listenAddr},
				service.Network{
					Properties: crg.NetworkProperties{
						Blockchain:          options.Blockchain,
						Network:             options.Network,
						AddrPrefix:          sdk.GetConfig().GetBech32AccountAddrPrefix(),
						SupportedOperations: client.SupportedOperations(),
					},
					Adapter: svc,
				})
			if err != nil {
				panic(err)
			}

			fmt.Printf("starting Rosetta API service at %s\n", listenAddr)

			err = s.Start()
			if err != nil {
				panic(err)
			}

			return nil
		},
	}

	cmd.Flags().String(flagBlockchain, "blockchain", "Application's name (e.g. Cosmos Hub)")
	cmd.Flags().String(flagListenAddr, "localhost:8080", "The address where Rosetta API will listen.")
	cmd.Flags().String(flagNetwork, "network", "Network's identifier (e.g. cosmos-hub-3, testnet-1, etc)")
	cmd.Flags().String(flagAppRPC, "localhost:1317", "Application's RPC endpoint.")
	cmd.Flags().String(flagTendermintRPC, "localhost:26657", "Tendermint's RPC endpoint.")
	cmd.Flags().Bool(flagOfflineMode, false, "Flag that forces the rosetta service to run in offline mode, some endpoints won't work.")

	return cmd
}

func getRosettaOptionsFromFlags(flags *flag.FlagSet) (Options, error) {
	blockchain, err := flags.GetString(flagBlockchain)
	if err != nil {
		return Options{}, fmt.Errorf("invalid blockchain value: %w", err)
	}

	network, err := flags.GetString(flagNetwork)
	if err != nil {
		return Options{}, fmt.Errorf("invalid network value: %w", err)
	}

	appRPC, err := flags.GetString(flagAppRPC)
	if err != nil {
		return Options{}, fmt.Errorf("invalid app rpc value: %w", err)
	}
	if !strings.HasPrefix(appRPC, "http://") {
		appRPC = fmt.Sprintf("http://%s", appRPC)
	}

	tendermintRPC, err := flags.GetString(flagTendermintRPC)
	if err != nil {
		return Options{}, fmt.Errorf("invalid tendermint rpc value: %w", err)
	}
	if !strings.HasPrefix(tendermintRPC, "tcp://") {
		tendermintRPC = fmt.Sprintf("tcp://%s", tendermintRPC)
	}

	offline, err := flags.GetBool(flagOfflineMode)
	if err != nil {
		return Options{}, fmt.Errorf("invalid offline value: %w", err)
	}

	return Options{
		AppEndpoint:        appRPC,
		TendermintEndpoint: tendermintRPC,
		Blockchain:         blockchain,
		Network:            network,
		OfflineMode:        offline,
	}, nil
}
