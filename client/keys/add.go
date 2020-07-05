package keys

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"

	bip39 "github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagInteractive = "interactive"
	flagRecover     = "recover"
	flagNoBackup    = "no-backup"
	flagCoinType    = "coin-type"
	flagAccount     = "account"
	flagIndex       = "index"
	flagMultisig    = "multisig"
	flagNoSort      = "nosort"
	flagHDPath      = "hd-path"
	flagKeyAlgo     = "algo"

	// DefaultKeyPass contains the default key password for genesis transactions
	DefaultKeyPass = "12345678"
)

// AddKeyCommand defines a keys command to add a generated or recovered private key to keybase.
func AddKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add an encrypted private key (either newly generated or recovered), encrypt it, and save to disk",
		Long: `Derive a new private key and encrypt to disk.
Optionally specify a BIP39 mnemonic, a BIP39 passphrase to further secure the mnemonic,
and a bip32 HD path to derive a specific account. The key will be stored under the given name
and encrypted with the given password. The only input that is required is the encryption password.

If run with -i, it will prompt the user for BIP44 path, BIP39 mnemonic, and passphrase.
The flag --recover allows one to recover a key from a seed passphrase.
If run with --dry-run, a key would be generated (or recovered) but not stored to the
local keystore.
Use the --pubkey flag to add arbitrary public keys to the keystore for constructing
multisig transactions.

You can add a multisig key by passing the list of key names you want the public
key to be composed of to the --multisig flag and the minimum number of signatures
required through --multisig-threshold. The keys are sorted by address, unless
the flag --nosort is set.
`,
		Args: cobra.ExactArgs(1),
		RunE: runAddCmd,
	}
	cmd.Flags().StringSlice(flagMultisig, nil, "Construct and store a multisig public key (implies --pubkey)")
	cmd.Flags().Uint(flagMultiSigThreshold, 1, "K out of N required signatures. For use in conjunction with --multisig")
	cmd.Flags().Bool(flagNoSort, false, "Keys passed to --multisig are taken in the order they're supplied")
	cmd.Flags().String(FlagPublicKey, "", "Parse a public key in bech32 format and save it to disk")
	cmd.Flags().BoolP(flagInteractive, "i", false, "Interactively prompt user for BIP39 passphrase and mnemonic")
	cmd.Flags().Bool(flags.FlagUseLedger, false, "Store a local reference to a private key on a Ledger device")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Bool(flags.FlagDryRun, false, "Perform action, but don't add key to local keystore")
	cmd.Flags().String(flagHDPath, "", "Manual HD Path derivation (overrides BIP44 config)")
	cmd.Flags().Uint32(flagCoinType, sdk.GetConfig().GetCoinType(), "coin type number for HD derivation")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")
	cmd.Flags().String(flagKeyAlgo, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")

	cmd.SetOut(cmd.OutOrStdout())
	cmd.SetErr(cmd.ErrOrStderr())

	return cmd
}

func getKeybase(transient bool, buf io.Reader) (keyring.Keyring, error) {
	if transient {
		return keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, viper.GetString(flags.FlagHome), buf)
	}

	return keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), buf)
}

func runAddCmd(cmd *cobra.Command, args []string) error {
	inBuf := bufio.NewReader(cmd.InOrStdin())
	kb, err := getKeybase(viper.GetBool(flags.FlagDryRun), inBuf)

	if err != nil {
		return err
	}

	return RunAddCmd(cmd, args, kb, inBuf)
}

/*
input
	- bip39 mnemonic
	- bip39 passphrase
	- bip44 path
	- local encryption password
output
	- armor encrypted private key (saved to file)
*/
func RunAddCmd(cmd *cobra.Command, args []string, kb keyring.Keyring, inBuf *bufio.Reader) error {
	var err error

	name := args[0]
	interactive := viper.GetBool(flagInteractive)
	showMnemonic := !viper.GetBool(flagNoBackup)

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(viper.GetString(flagKeyAlgo), keyringAlgos)
	if err != nil {
		return err
	}

	if !viper.GetBool(flags.FlagDryRun) {
		_, err = kb.Key(name)
		if err == nil {
			// account exists, ask for user confirmation
			response, err2 := input.GetConfirmation(fmt.Sprintf("override the existing name %s", name), inBuf, cmd.ErrOrStderr())
			if err2 != nil {
				return err2
			}

			if !response {
				return errors.New("aborted")
			}

			err2 = kb.Delete(name)
			if err2 != nil {
				return err2
			}
		}

		multisigKeys := viper.GetStringSlice(flagMultisig)
		if len(multisigKeys) != 0 {
			var pks []crypto.PubKey

			multisigThreshold := viper.GetInt(flagMultiSigThreshold)
			if err := validateMultisigThreshold(multisigThreshold, len(multisigKeys)); err != nil {
				return err
			}

			for _, keyname := range multisigKeys {
				k, err := kb.Key(keyname)
				if err != nil {
					return err
				}

				pks = append(pks, k.GetPubKey())
			}

			// Handle --nosort
			if !viper.GetBool(flagNoSort) {
				sort.Slice(pks, func(i, j int) bool {
					return bytes.Compare(pks[i].Address(), pks[j].Address()) < 0
				})
			}

			pk := multisig.NewPubKeyMultisigThreshold(multisigThreshold, pks)
			if _, err := kb.SaveMultisig(name, pk); err != nil {
				return err
			}

			cmd.PrintErrf("Key %q saved to disk.\n", name)

			return nil
		}
	}

	if viper.GetString(FlagPublicKey) != "" {
		pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, viper.GetString(FlagPublicKey))
		if err != nil {
			return err
		}

		if _, err := kb.SavePubKey(name, pk, algo.Name()); err != nil {
			return err
		}

		return nil
	}

	coinType := uint32(viper.GetInt(flagCoinType))
	account := uint32(viper.GetInt(flagAccount))
	index := uint32(viper.GetInt(flagIndex))
	hdPath := viper.GetString(flagHDPath)

	if len(hdPath) == 0 {
		hdPath = hd.CreateHDPath(coinType, account, index).String()
	} else if viper.GetBool(flags.FlagUseLedger) {
		return errors.New("cannot set custom bip32 path with ledger")
	}

	// If we're using ledger, only thing we need is the path and the bech32 prefix.
	if viper.GetBool(flags.FlagUseLedger) {
		bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
		info, err := kb.SaveLedgerKey(name, hd.Secp256k1, bech32PrefixAccAddr, coinType, account, index)

		if err != nil {
			return err
		}

		return printCreate(cmd, info, false, "")
	}

	// Get bip39 mnemonic
	var mnemonic, bip39Passphrase string

	if interactive || viper.GetBool(flagRecover) {
		bip39Message := "Enter your bip39 mnemonic"
		if !viper.GetBool(flagRecover) {
			bip39Message = "Enter your bip39 mnemonic, or hit enter to generate one."
		}

		mnemonic, err = input.GetString(bip39Message, inBuf)
		if err != nil {
			return err
		}

		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("invalid mnemonic")
		}
	}

	if len(mnemonic) == 0 {
		// read entropy seed straight from crypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			return err
		}
	}

	// override bip39 passphrase
	if interactive {
		bip39Passphrase, err = input.GetString(
			"Enter your bip39 passphrase. This is combined with the mnemonic to derive the seed. "+
				"Most users should just hit enter to use the default, \"\"", inBuf)
		if err != nil {
			return err
		}

		// if they use one, make them re-enter it
		if len(bip39Passphrase) != 0 {
			p2, err := input.GetString("Repeat the passphrase:", inBuf)
			if err != nil {
				return err
			}

			if bip39Passphrase != p2 {
				return errors.New("passphrases don't match")
			}
		}
	}

	info, err := kb.NewAccount(name, mnemonic, bip39Passphrase, hdPath, algo)
	if err != nil {
		return err
	}

	// Recover key from seed passphrase
	if viper.GetBool(flagRecover) {
		// Hide mnemonic from output
		showMnemonic = false
		mnemonic = ""
	}

	return printCreate(cmd, info, showMnemonic, mnemonic)
}

func printCreate(cmd *cobra.Command, info keyring.Info, showMnemonic bool, mnemonic string) error {
	output := viper.Get(cli.OutputFlag)

	switch output {
	case OutputFormatText:
		cmd.PrintErrln()
		printKeyInfo(cmd.OutOrStdout(), info, keyring.Bech32KeyOutput)

		// print mnemonic unless requested not to.
		if showMnemonic {
			fmt.Fprintln(cmd.ErrOrStderr(), "\n**Important** write this mnemonic phrase in a safe place.")
			fmt.Fprintln(cmd.ErrOrStderr(), "\n**Important** write this mnemonic phrase in a safe place.")
			fmt.Fprintln(cmd.ErrOrStderr(), "It is the only way to recover your account if you ever forget your password.")
			fmt.Fprintln(cmd.ErrOrStderr(), "")
			fmt.Fprintln(cmd.ErrOrStderr(), mnemonic)
		}
	case OutputFormatJSON:
		out, err := keyring.Bech32KeyOutput(info)
		if err != nil {
			return err
		}

		if showMnemonic {
			out.Mnemonic = mnemonic
		}

		jsonString, err := KeysCdc.MarshalJSON(out)
		if err != nil {
			return err
		}

		cmd.Println(string(jsonString))

	default:
		return fmt.Errorf("invalid output format %s", output)
	}

	return nil
}
