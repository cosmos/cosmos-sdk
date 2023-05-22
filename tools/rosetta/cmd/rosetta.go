package cmd

import (
	"fmt"
	"os"
	"plugin"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/rosetta"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := rosetta.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			protoCodec, ok := cdc.(*codec.ProtoCodec)
			if !ok {
				return fmt.Errorf("exoected *codec.ProtoMarshaler, got: %T", cdc)
			}
			conf.WithCodec(ir, protoCodec)

			// load module
			// 1. open the so file to load the symbols
			plug, err := plugin.Open("./plugins/osmosis.so")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// 2. look up a symbol (an exported function or variable)
			initZone, err := plug.Lookup("InitZone")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			initZone.(func())()

			rosettaSrv, err := rosetta.ServerFromConfig(conf)
			if err != nil {
				fmt.Printf("[Rosetta]- Error while creating server: %s", err.Error())
				return err
			}
			return rosettaSrv.Start()
		},
	}
	rosetta.SetFlags(cmd.Flags())

	return cmd
}
