package keys

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	cmd.Flags().String(FlagBechPrefix, sdk.PrefixAccount, "The Bech32 prefix encoding for a key (acc|val|cons)")
	cmd.Flags().BoolP(FlagAddress, "a", false, "Output the address only (overrides --output)")
	cmd.Flags().BoolP(FlagPublicKey, "p", false, "Output the public key only (overrides --output)")
	cmd.Flags().BoolP(FlagDevice, "d", false, "Output the address in a ledger device")
	cmd.Flags().Uint(flagMultiSigThreshold, 1, "K out of N required signatures")

	return cmd
}

func runShowCmd(cmd *cobra.Command, args []string) (err error) {
	var info keyring.Info

	kb, err := keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), cmd.InOrStdin())
	if err != nil {
		return err
	}
	if len(args) == 1 {
		info, err = fetchKey(kb, args[0])
		if err != nil {
			return fmt.Errorf("%s is not a valid name or address: %v", args[0], err)
		}
	} else {
		pks := make([]tmcrypto.PubKey, len(args))
		for i, keyref := range args {
			info, err := fetchKey(kb, keyref)
			if err != nil {
				return fmt.Errorf("%s is not a valid name or address: %v", keyref, err)
			}

			pks[i] = info.GetPubKey()
		}

		multisigThreshold := viper.GetInt(flagMultiSigThreshold)
		err = validateMultisigThreshold(multisigThreshold, len(args))
		if err != nil {
			return err
		}

		multikey := multisig.NewPubKeyMultisigThreshold(multisigThreshold, pks)
		info = keyring.NewMultiInfo(defaultMultiSigKeyName, multikey)
	}

	isShowAddr := viper.GetBool(FlagAddress)
	isShowPubKey := viper.GetBool(FlagPublicKey)
	isShowDevice := viper.GetBool(FlagDevice)

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

	bechKeyOut, err := getBechKeyOut(viper.GetString(FlagBechPrefix))
	if err != nil {
		return err
	}

	switch {
	case isShowAddr:
		printKeyAddress(cmd.OutOrStdout(), info, bechKeyOut)
	case isShowPubKey:
		printPubKey(cmd.OutOrStdout(), info, bechKeyOut)
	default:
		printKeyInfo(cmd.OutOrStdout(), info, bechKeyOut)
	}

	if isShowDevice {
		if isShowPubKey {
			return fmt.Errorf("the device flag (-d) can only be used for addresses not pubkeys")
		}
		if viper.GetString(FlagBechPrefix) != "acc" {
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

		return crypto.LedgerShowAddress(*hdpath, info.GetPubKey(), sdk.GetConfig().GetBech32AccountAddrPrefix())
	}

	return nil
}

func fetchKey(kb keyring.Keyring, keyref string) (keyring.Info, error) {
	info, err := kb.Key(keyref)
	if err != nil {
		accAddr, err := sdk.AccAddressFromBech32(keyref)
		if err != nil {
			return info, err
		}

		info, err = kb.KeyByAddress(accAddr)
		if err != nil {
			return info, errors.New("key not found")
		}
	}
	return info, nil
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
		return keyring.Bech32KeyOutput, nil
	case sdk.PrefixValidator:
		return keyring.Bech32ValKeyOutput, nil
	case sdk.PrefixConsensus:
		return keyring.Bech32ConsKeyOutput, nil
	}

	return nil, fmt.Errorf("invalid Bech32 prefix encoding provided: %s", bechPrefix)
}
