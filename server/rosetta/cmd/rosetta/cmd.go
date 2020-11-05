package main

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"net/http"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "cosmos rosetta API implementation server",
	}
	cmd.AddCommand(startCmd())
	return cmd
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Info("finding configuration")
			config, err := FindConfig(cmd.Flags())
			if err != nil {
				return err
			}
			klog.Infof("configuration found: %#v", config)
			svc, err := RetryRosettaFromConfig(config)
			if err != nil {
				return err
			}
			klog.Infof("cosmos rosetta adapter built")
			router, err := rosetta.NewRouter(&types.NetworkIdentifier{
				Blockchain:           config.Blockchain,
				Network:              config.Network,
				SubNetworkIdentifier: nil,
			}, svc)
			klog.Infof("http router correctly instantiated")
			if err != nil {
				return err
			}
			klog.Infof("listening and serving at: %s", config.Addr)
			return http.ListenAndServe(config.Addr, router)
		},
	}

	SetConfigFlags(cmd.Flags())
	return cmd
}
