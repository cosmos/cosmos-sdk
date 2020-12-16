package server

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/cosmos-rosetta-gateway/service"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/config"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc codec.Marshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			if protoCodec, ok := cdc.(*codec.ProtoCodec); ok {
				conf.WithCodec(ir, protoCodec)
			}
			if err := conf.Validate(); err != nil {
				return err
			}

			adapter, client, err := config.RetryRosettaFromConfig(conf)
			if err != nil {
				return err
			}

			svc, err := service.New(
				service.Options{ListenAddress: conf.Addr},
				rosetta.NewNetwork(conf.NetworkIdentifier(), adapter, client),
			)
			if err != nil {
				return err
			}
			return svc.Start()
		},
	}
	config.SetFlags(cmd.Flags())

	return cmd
}
