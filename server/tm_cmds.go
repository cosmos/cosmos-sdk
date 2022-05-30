package server

// DONTCOVER

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	cfg "github.com/tendermint/tendermint/config"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/scripts/keymigrate"
	"github.com/tendermint/tendermint/scripts/scmigrate"
	tversion "github.com/tendermint/tendermint/version"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ShowNodeIDCmd - ported from Tendermint, dump node ID to stdout
func ShowNodeIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-node-id",
		Short: "Show this node's ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			nodeKey, err := p2p.LoadNodeKey(cfg.NodeKeyFile())
			if err != nil {
				return err
			}
			fmt.Println(nodeKey.ID())
			return nil
		},
	}
}

// ShowValidatorCmd - ported from Tendermint, show this node's validator info
func ShowValidatorCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show-validator",
		Short: "Show this node's tendermint validator info",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			pk, err := privValidator.GetPubKey()
			if err != nil {
				return err
			}
			sdkPK, err := cryptocodec.FromTmPubKeyInterface(pk)
			if err != nil {
				return err
			}
			clientCtx := client.GetClientContextFromCmd(cmd)
			bz, err := clientCtx.Codec.MarshalInterfaceJSON(sdkPK)
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
		},
	}

	return &cmd
}

// ShowAddressCmd - show this node's validator address
func ShowAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-address",
		Short: "Shows this node's tendermint validator consensus address",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			valConsAddr := (sdk.ConsAddress)(privValidator.GetAddress())
			fmt.Println(valConsAddr.String())
			return nil
		},
	}

	return cmd
}

// VersionCmd prints tendermint and ABCI version numbers.
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print tendermint libraries' version",
		Long: `Print protocols' and libraries' version numbers
against which this app has been compiled.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			bs, err := yaml.Marshal(&struct {
				Tendermint    string
				ABCI          string
				BlockProtocol uint64
				P2PProtocol   uint64
			}{
				Tendermint:    tversion.TMCoreSemVer,
				ABCI:          tversion.ABCIVersion,
				BlockProtocol: tversion.BlockProtocol,
				P2PProtocol:   tversion.P2PProtocol,
			})
			if err != nil {
				return err
			}

			fmt.Println(string(bs))
			return nil
		},
	}
}

// makeKeyMigrateCmd is ported from tendermint's key-migrate command, but
// uses the SDK's own server.Context.
// ref: https://github.com/tendermint/tendermint/blob/master/UPGRADING.md#database-key-format-changes
func makeKeyMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-migrate",
		Short: "Run Tendermint database key migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			serverCtx := GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			contexts := []string{
				// this is ordered to put the
				// (presumably) biggest/most important
				// subsets first.
				"blockstore",
				"state",
				"peerstore",
				"tx_index",
				"evidence",
				"light",
			}

			for idx, dbctx := range contexts {
				serverCtx.Logger.Info("beginning a key migration",
					"dbctx", dbctx,
					"num", idx+1,
					"total", len(contexts),
				)

				db, err := cfg.DefaultDBProvider(&cfg.DBContext{
					ID:     dbctx,
					Config: config,
				})

				if err != nil {
					return fmt.Errorf("constructing database handle: %w", err)
				}

				if err = keymigrate.Migrate(ctx, db); err != nil {
					return fmt.Errorf("running migration for context %q: %w",
						dbctx, err)
				}

				if dbctx == "blockstore" {
					if err := scmigrate.Migrate(ctx, db); err != nil {
						return fmt.Errorf("running seen commit migration: %w", err)

					}
				}
			}

			serverCtx.Logger.Info("completed database migration successfully")

			return nil
		},
	}

	return cmd
}
