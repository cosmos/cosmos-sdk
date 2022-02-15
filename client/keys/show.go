package keys

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/ledger"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerr "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// FlagAddress is the flag for the user's address on the command line.
	FlagAddress = "address"
	// FlagPublicKey represents the user's public key on the command line.
	FlagPublicKey = "pubkey"
	// FlagBechPrefix defines a desired Bech32 prefix encoding for a key.
	FlagBechPrefix = "bech"
	// FlagDevice indicates that the information should be shown in the device
	FlagDevice = "device"

	flagMultiSigThreshold = "multisig-threshold"

	defaultMultiSigKeyName = "multi"
)

// ShowKeysCmd shows key information for a given key name.
func ShowKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [name_or_address [name_or_address...]]",
		Short: "Retrieve key information by name or address",
		Long: `Display keys details. If multiple names or addresses are provided,
then an ephemeral multisig key will be created under the name "multi"
consisting of all the keys provided by name and multisig threshold.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runShowCmd,
	}
	f := cmd.Flags()
	f.String(FlagBechPrefix, sdk.PrefixAccount, "The Bech32 prefix encoding for a key (acc|val|cons)")
	f.BoolP(FlagAddress, "a", false, "Output the address only (overrides --output)")
	f.BoolP(FlagPublicKey, "p", false, "Output the public key only (overrides --output)")
	f.BoolP(FlagDevice, "d", false, "Output the address in a ledger device")
	f.Int(flagMultiSigThreshold, 1, "K out of N required signatures")

	return cmd
}

func runShowCmd(cmd *cobra.Command, args []string) (err error) {
	var info keyring.Info
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		info, err = fetchKey(clientCtx.Keyring, args[0])
		if err != nil {
			return fmt.Errorf("%s is not a valid name or address: %v", args[0], err)
		}
	} else {
		pks := make([]cryptotypes.PubKey, len(args))
		for i, keyref := range args {
			info, err := fetchKey(clientCtx.Keyring, keyref)
			if err != nil {
				return fmt.Errorf("%s is not a valid name or address: %v", keyref, err)
			}

			pks[i] = info.GetPubKey()
		}

		multisigThreshold, _ := cmd.Flags().GetInt(flagMultiSigThreshold)
		err = validateMultisigThreshold(multisigThreshold, len(args))
		if err != nil {
			return err
		}

		multikey := multisig.NewLegacyAminoPubKey(multisigThreshold, pks)
		info, err = keyring.NewMultiInfo(defaultMultiSigKeyName, multikey)
		if err != nil {
			return err
		}
	}

	isShowAddr, _ := cmd.Flags().GetBool(FlagAddress)
	isShowPubKey, _ := cmd.Flags().GetBool(FlagPublicKey)
	isShowDevice, _ := cmd.Flags().GetBool(FlagDevice)

	isOutputSet := false
	tmp := cmd.Flag(cli.OutputFlag)
	if tmp != nil {
		isOutputSet = tmp.Changed
	}

	if isShowAddr && isShowPubKey {
		return errors.New("cannot use both --address and --pubkey at once")
	}

	if isOutputSet && (isShowAddr || isShowPubKey) {
		return errors.New("cannot use --output with --address or --pubkey")
	}

	bechPrefix, _ := cmd.Flags().GetString(FlagBechPrefix)
	bechKeyOut, err := getBechKeyOut(bechPrefix)
	if err != nil {
		return err
	}

	if isOutputSet {
		clientCtx.OutputFormat, _ = cmd.Flags().GetString(cli.OutputFlag)
	}

	switch {
	case isShowAddr, isShowPubKey:
		ko, err := bechKeyOut(info)
		if err != nil {
			return err
		}
		out := ko.Address
		if isShowPubKey {
			out = ko.PubKey
		}
		fmt.Fprintln(cmd.OutOrStdout(), out)
	default:
		printKeyInfo(cmd.OutOrStdout(), info, bechKeyOut, clientCtx.OutputFormat)
	}

	if isShowDevice {
		if isShowPubKey {
			return fmt.Errorf("the device flag (-d) can only be used for addresses not pubkeys")
		}
		if bechPrefix != "acc" {
			return fmt.Errorf("the device flag (-d) can only be used for accounts")
		}

		// Override and show in the device
		if info.GetType() != keyring.TypeLedger {
			return fmt.Errorf("the device flag (-d) can only be used for accounts stored in devices")
		}

		hdpath, err := info.GetPath()
		if err != nil {
			return nil
		}

		return ledger.ShowAddress(*hdpath, info.GetPubKey(), sdk.GetConfig().GetBech32AccountAddrPrefix())
	}

	return nil
}

func fetchKey(kb keyring.Keyring, keyref string) (keyring.Info, error) {
	// firstly check if the keyref is a key name of a key registered in a keyring.
	info, err := kb.Key(keyref)
	// if the key is not there or if we have a problem with a keyring itself then we move to a
	// fallback: searching for key by address.
	if err == nil || !sdkerr.IsOf(err, sdkerr.ErrIO, sdkerr.ErrKeyNotFound) {
		return info, err
	}
	accAddr, err := sdk.AccAddressFromBech32(keyref)
	if err != nil {
		return info, err
	}

	info, err = kb.KeyByAddress(accAddr)
	return info, sdkerr.Wrap(err, "Invalid key")
}

func validateMultisigThreshold(k, nKeys int) error {
	if k <= 0 {
		return fmt.Errorf("threshold must be a positive integer")
	}
	if nKeys < k {
		return fmt.Errorf(
			"threshold k of n multisignature: %d < %d", nKeys, k)
	}
	return nil
}

func getBechKeyOut(bechPrefix string) (bechKeyOutFn, error) {
	switch bechPrefix {
	case sdk.PrefixAccount:
		return keyring.MkAccKeyOutput, nil
	case sdk.PrefixValidator:
		return keyring.MkValKeyOutput, nil
	case sdk.PrefixConsensus:
		return keyring.MkConsKeyOutput, nil
	}

	return nil, fmt.Errorf("invalid Bech32 prefix encoding provided: %s", bechPrefix)
}
