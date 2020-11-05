package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/config"
	"github.com/spf13/cobra"
	"net/http"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc *codec.ProtoCodec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get config
			conf, err := config.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			// add codec settings to config
			conf.WithCodec(ir, cdc)
			// validate config
			if err := conf.Validate(); err != nil {
				return err
			}
			// instantiate a new rosetta service
			svc, err := config.RetryRosettaFromConfig(conf)
			if err != nil {
				return err
			}
			// create the router
			router, err := rosetta.NewRouter(conf.NetworkIdentifier(), svc)
			if err != nil {
				return err
			}
			return http.ListenAndServe(conf.Addr, router)
		},
	}
	config.SetFlags(cmd.Flags(), config.DisableFileFlag())
	return cmd
}
