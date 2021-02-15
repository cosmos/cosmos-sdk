package server

// DONTCOVER

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	tversion "github.com/tendermint/tendermint/version"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
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
			valPubKey, err := privValidator.GetPubKey()
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString(cli.OutputFlag)
			if strings.ToLower(output) == "json" {
				return printlnJSON(valPubKey)
			}

			pubkey, err := cryptocodec.FromTmPubKeyInterface(valPubKey)
			if err != nil {
				return err
			}
			pubkeyBech32, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pubkey)
			if err != nil {
				return err
			}

			fmt.Println(pubkeyBech32)
			return nil
		},
	}

	cmd.Flags().StringP(cli.OutputFlag, "o", "text", "Output format (text|json)")
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

			output, _ := cmd.Flags().GetString(cli.OutputFlag)
			if strings.ToLower(output) == "json" {
				return printlnJSON(valConsAddr)
			}

			fmt.Println(valConsAddr.String())
			return nil
		},
	}

	cmd.Flags().StringP(cli.OutputFlag, "o", "text", "Output format (text|json)")
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

func printlnJSON(v interface{}) error {
	cdc := codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(cdc)

	marshalled, err := cdc.MarshalJSON(v)
	if err != nil {
		return err
	}

	fmt.Println(string(marshalled))
	return nil
}

// UnsafeResetAllCmd - extension of the tendermint command, resets initialization
func UnsafeResetAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unsafe-reset-all",
		Short: "Resets the blockchain database, removes address book files, and resets data/priv_validator_state.json to the genesis state",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			tcmd.ResetAll(cfg.DBDir(), cfg.P2P.AddrBookFile(), cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile(), serverCtx.Logger)
			return nil
		},
	}
}
