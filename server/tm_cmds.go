package server

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/codec"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	tversion "github.com/tendermint/tendermint/version"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ShowNodeIDCmd - ported from Tendermint, dump node ID to stdout
func ShowNodeIDCmd(ctx *Context) *cobra.Command {
	return &cobra.Command{
		Use:   "show-node-id",
		Short: "Show this node's ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.Config
			nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
			if err != nil {
				return err
			}
			fmt.Println(nodeKey.ID())
			return nil
		},
	}
}

// ShowValidator - ported from Tendermint, show this node's validator info
func ShowValidatorCmd(ctx *Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "show-validator",
		Short: "Show this node's tendermint validator info",
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg := ctx.Config
			UpgradeOldPrivValFile(cfg)
			privValidator := pvm.LoadOrGenFilePV(
				cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			valPubKey := privValidator.GetPubKey()

			if viper.GetBool(client.FlagJson) {
				return printlnJSON(valPubKey)
			}

			pubkey, err := sdk.Bech32ifyConsPub(valPubKey)
			if err != nil {
				return err
			}

			fmt.Println(pubkey)
			return nil
		},
	}
	cmd.Flags().Bool(client.FlagJson, false, "get machine parseable output")
	return &cmd
}

// ShowAddressCmd - show this node's validator address
func ShowAddressCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-address",
		Short: "Shows this node's tendermint validator consensus address",
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg := ctx.Config
			UpgradeOldPrivValFile(cfg)
			privValidator := pvm.LoadOrGenFilePV(
				cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			valConsAddr := (sdk.ConsAddress)(privValidator.GetAddress())

			if viper.GetBool(client.FlagJson) {
				return printlnJSON(valConsAddr)
			}

			fmt.Println(valConsAddr.String())
			return nil
		},
	}

	cmd.Flags().Bool(client.FlagJson, false, "get machine parseable output")
	return cmd
}

// VersionCmd prints tendermint and ABCI version numbers.
func VersionCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print tendermint libraries' version",
		Long: `Print Tendermint's and ABCI's version numbers
against which this app has been compiled.
`,
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Printf(`tendermint: %s
ABCI: %s
`, tversion.Version, tversion.ABCIVersion)

			return nil
		},
	}
	return cmd
}

func printlnJSON(v interface{}) error {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	marshalled, err := cdc.MarshalJSON(v)
	if err != nil {
		return err
	}
	fmt.Println(string(marshalled))
	return nil
}

// UnsafeResetAllCmd - extension of the tendermint command, resets initialization
func UnsafeResetAllCmd(ctx *Context) *cobra.Command {
	return &cobra.Command{
		Use:   "unsafe-reset-all",
		Short: "Resets the blockchain database, removes address book files, and resets priv_validator.json to the genesis state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.Config
			tcmd.ResetAll(cfg.DBDir(), cfg.P2P.AddrBookFile(), cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile(), ctx.Logger)
			return nil
		},
	}
}
