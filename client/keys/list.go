package keys

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

const (
	flagListNames     = "list-names"
	flagListAddresses = "list-addresses"
	flagListPubKeys   = "list-pubkeys"
)

// ListKeysCmd lists all keys in the key store.
func ListKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys",
		Long: `Return a list of all public keys stored by this key manager
along with their associated name and address.`,
		RunE: runListCmd,
	}

	cmd.Flags().BoolP(flagListNames, "n", false, "List names only (not compatible with --output)")
	cmd.Flags().BoolP(flagListAddresses, "a", false, "List addresses only (not compatible with --output)")
	cmd.Flags().BoolP(flagListPubKeys, "p", false, "List public keys only (not compatible with --output)")
	return cmd
}

func runListCmd(cmd *cobra.Command, _ []string) error {
	isShowAddr, _ := cmd.Flags().GetBool(flagListAddresses)
	isShowPubKey, _ := cmd.Flags().GetBool(flagListPubKeys)
	isShowName, _ := cmd.Flags().GetBool(flagListNames)
	isOutputSet := false
	tmp := cmd.Flag(flags.FlagOutput)
	if tmp != nil {
		isOutputSet = tmp.Changed
	}
	if isOutputSet && (isShowAddr || isShowPubKey || isShowName) {
		return errors.New("cannot use --output with --list-addresses or --list-pubkeys or --list-names")
	}
	switch {
	case isShowAddr && isShowPubKey:
		return errors.New("cannot use both --list-addresses and --list-pubkeys at once")
	case isShowAddr && isShowName:
		return errors.New("cannot use both --list-addresses and --list-names at once")
	case isShowPubKey && isShowName:
		return errors.New("cannot use both --list-pubkeys and --list-names at once")
	}

	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	records, err := clientCtx.Keyring.List()
	if err != nil {
		return err
	}

	if len(records) == 0 && clientCtx.OutputFormat == flags.OutputFormatText {
		cmd.Println("No records were found in keyring")
		return nil
	}
	switch {
	case isShowName:
		for _, k := range records {
			cmd.Println(k.Name)
		}
	case isShowPubKey:
		kos, err := MkAccKeysOutput(records, clientCtx.AddressCodec)
		if err != nil {
			return err
		}
		for _, k := range kos {
			cmd.Println(k.PubKey)
		}
	case isShowAddr:
		kos, err := MkAccKeysOutput(records, clientCtx.AddressCodec)
		if err != nil {
			return err
		}
		for _, k := range kos {
			cmd.Println(k.Address)
		}
	default:
		return printKeyringRecords(clientCtx, cmd.OutOrStdout(), records, clientCtx.OutputFormat)
	}

	return nil
}

// ListKeyTypesCmd lists all key types.
func ListKeyTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-key-types",
		Short: "List all key types",
		Long:  `Return a list of all supported key types (also known as algos)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			cmd.Println("Supported key types/algos:")
			keyring, _ := clientCtx.Keyring.SupportedAlgorithms()
			cmd.Printf("%+q\n", keyring)
			return nil
		},
	}
}
