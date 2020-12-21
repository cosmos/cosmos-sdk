package server

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/server/rosetta"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc codec.Marshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := rosetta.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			if protoCodec, ok := cdc.(*codec.ProtoCodec); ok {
				conf.WithCodec(ir, protoCodec)
			}
			rosettaSrv, err := rosetta.ServerFromConfig(conf)
			if err != nil {
				return err
			}
			return rosettaSrv.Start()
		},
	}
	rosetta.SetFlags(cmd.Flags())

	return cmd
}
