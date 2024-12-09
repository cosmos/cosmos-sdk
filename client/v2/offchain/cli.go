package offchain

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/autocli/config"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/client/v2/broadcast/comet"
	clientcontext "cosmossdk.io/client/v2/context"
	v2flags "cosmossdk.io/client/v2/internal/flags"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	flagEncoding   = "encoding"
	flagFileFormat = "file-format"
	flagBech32     = "bech32"
)

// OffChain off-chain utilities.
func OffChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "off-chain",
		Short: "Off-chain utilities.",
		Long:  `Utilities for off-chain data.`,
	}

	cmd.AddCommand(
		SignFile(),
		VerifyFile(),
	)

	flags.AddKeyringFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().String(flagBech32, "cosmos", "address bech32 prefix")
	return cmd
}

// SignFile signs a file with a key.
func SignFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-file <keyName> <fileName>",
		Short: "Sign a file.",
		Long:  "Sign a file using a given key.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ir := types.NewInterfaceRegistry()
			cryptocodec.RegisterInterfaces(ir)
			cdc := codec.NewProtoCodec(ir)

			c, err := config.CreateClientConfigFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			keyringBackend := c.KeyringBackend
			if !cmd.Flags().Changed(v2flags.FlagKeyringBackend) {
				_ = cmd.Flags().Set(v2flags.FlagKeyringBackend, keyringBackend)
			}

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			encoding, _ := cmd.Flags().GetString(flagEncoding)
			outputFormat, _ := cmd.Flags().GetString(v2flags.FlagOutput)
			outputFile, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			signMode, _ := cmd.Flags().GetString(flags.FlagSignMode)
			bech32Prefix, _ := cmd.Flags().GetString(flagBech32)

			ac := address.NewBech32Codec(bech32Prefix)
			k, err := keyring.NewKeyringFromFlags(cmd.Flags(), ac, cmd.InOrStdin(), cdc)
			if err != nil {
				return err
			}

			// off-chain does not need to query any information
			conn, err := comet.NewCometBFTBroadcaster("", comet.BroadcastSync, cdc)
			if err != nil {
				return err
			}

			ctx := clientcontext.Context{
				Flags:                 cmd.Flags(),
				AddressCodec:          ac,
				ValidatorAddressCodec: address.NewBech32Codec(sdk.GetBech32PrefixValAddr(bech32Prefix)),
				Cdc:                   cdc,
				Keyring:               k,
			}

			signedTx, err := Sign(ctx, bz, conn, args[0], encoding, signMode, outputFormat)
			if err != nil {
				return err
			}

			if outputFile != "" {
				fp, err := os.OpenFile(filepath.Clean(outputFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
				if err != nil {
					return err
				}
				cmd.SetOut(fp)
			}

			cmd.Println(signedTx)
			return nil
		},
	}

	cmd.Flags().String(v2flags.FlagOutput, "json", "Choose an output format for the tx (json|text")
	cmd.Flags().String(flagEncoding, "no-encoding", "Choose an encoding method for the file content to be added as msg data (no-encoding|base64|hex)")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.PersistentFlags().String(flags.FlagSignMode, "direct", "Choose sign mode (direct|amino-json)")
	return cmd
}

// VerifyFile verifies given file with given key.
func VerifyFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-file <signedFileName>",
		Short: "Verify a file.",
		Long:  "Verify a previously signed file with the given key.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ir := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(ir)

			bz, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			fileFormat, _ := cmd.Flags().GetString(flagFileFormat)
			bech32Prefix, _ := cmd.Flags().GetString(flagBech32)

			ac := address.NewBech32Codec(bech32Prefix)

			ctx := clientcontext.Context{
				Flags:                 cmd.Flags(),
				AddressCodec:          ac,
				ValidatorAddressCodec: address.NewBech32Codec(sdk.GetBech32PrefixValAddr(bech32Prefix)),
				Cdc:                   cdc,
			}

			err = Verify(ctx, bz, fileFormat)
			if err == nil {
				cmd.Println("Verification OK!")
			}
			return err
		},
	}

	cmd.Flags().String(flagFileFormat, "json", "Choose what's the file format to be verified (json|text)")
	return cmd
}
