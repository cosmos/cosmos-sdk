package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/config"
	"github.com/spf13/cobra"
	"github.com/tendermint/cosmos-rosetta-gateway/service"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc codec.Marshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get config
			conf, err := config.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			// if the provided interface registry and codec
			// are valid then use them
			if protoCodec, ok := cdc.(*codec.ProtoCodec); ok {
				conf.WithCodec(ir, protoCodec)
			}
			// validate config
			if err := conf.Validate(); err != nil {
				return err
			}
			// instantiate a new rosetta service
			adapter, err := config.RetryRosettaFromConfig(conf)
			if err != nil {
				return err
			}
			// create the router
			svc, err := service.New(
				service.Options{ListenAddress: conf.Addr},
				rosetta.NewNetwork(conf.NetworkIdentifier(), adapter),
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
