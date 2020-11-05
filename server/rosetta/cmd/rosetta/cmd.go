package main

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/config"
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
			conf, err := config.Find(cmd.Flags())
			if err != nil {
				return err
			}
			klog.Infof("configuration found: %#v", conf)
			svc, err := config.RetryRosettaFromConfig(conf)
			if err != nil {
				return err
			}
			klog.Infof("cosmos rosetta adapter built")
			router, err := rosetta.NewRouter(&types.NetworkIdentifier{
				Blockchain:           conf.Blockchain,
				Network:              conf.Network,
				SubNetworkIdentifier: nil,
			}, svc)
			klog.Infof("http router correctly instantiated")
			if err != nil {
				return err
			}
			klog.Infof("listening and serving at: %s", conf.Addr)
			return http.ListenAndServe(conf.Addr, router)
		},
	}

	config.SetFlags(cmd.Flags())
	return cmd
}
