package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	rosettacfg "github.com/cosmos/cosmos-sdk/server/rosetta/config"
	"github.com/spf13/cobra"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc codec.Marshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := rosettacfg.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			if protoCodec, ok := cdc.(*codec.ProtoCodec); ok {
				conf.WithCodec(ir, protoCodec)
			}
			rosettaSrv, err := rosettacfg.HandlerFromConfig(conf)
			if err != nil {
				return err
			}
			return rosettaSrv.Start()
		},
	}
	rosettacfg.SetFlags(cmd.Flags())

	return cmd
}
