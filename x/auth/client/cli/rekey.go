package cli

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const flagProof = "proof"

// NewTxCmd returns the custom tx commands of the auth module.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSignRekeyProofCmd(),
		NewChangePubKeyCmd(),
	)

	return txCmd
}

// NewSignRekeyProofCmd returns a command that signs the proof of possession a
// MsgChangePubKey needs from its new pubkey.
func NewSignRekeyProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-rekey-proof [account-address] [new-pubkey-json]",
		Short: "Sign a proof of possession of a new pubkey for an account",
		Long: `Sign the proof of possession that a change-pubkey transaction needs from the
new pubkey of [account-address]. The --from key must be the new pubkey itself or,
if the new pubkey is a multisig, one of its direct members. Each member of a
multisig signs separately; pass the outputs to change-pubkey --proof.

The proof binds the chain id and the account's number and address, so it cannot
be reused for another account or chain. It is always signed in amino-json mode,
which Ledger devices support. In offline mode, --account-number is required.`,
		Example: fmt.Sprintf(`%s tx auth sign-rekey-proof cosmos1... '{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[...]}' --from member1 > proof1.json`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address, newPk, err := parseRekeyArgs(clientCtx, args)
			if err != nil {
				return err
			}

			if clientCtx.FromName == "" {
				return errors.New("a --from key is required to sign the proof")
			}
			rec, err := clientCtx.Keyring.Key(clientCtx.FromName)
			if err != nil {
				return err
			}
			signerPk, err := rec.GetPubKey()
			if err != nil {
				return err
			}
			if !isRekeyProofSigner(newPk, signerPk) {
				return fmt.Errorf("key %s is not the new pubkey or a member of it", clientCtx.FromName)
			}

			accNum, err := rekeyAccountNumber(cmd, clientCtx, address)
			if err != nil {
				return err
			}
			if clientCtx.ChainID == "" {
				return errors.New("set the chain id with either the --chain-id flag or config file")
			}

			anyPk, err := codectypes.NewAnyWithValue(newPk)
			if err != nil {
				return err
			}
			doc := types.ChangePubKeyProofDoc{
				ChainId:       clientCtx.ChainID,
				AccountNumber: accNum,
				Address:       address,
				NewPubKey:     anyPk,
			}
			signBytes, err := types.ChangePubKeyProofSignBytes(clientCtx.TxConfig.SigningContext().AddressCodec(), doc, newPk)
			if err != nil {
				return err
			}

			// The chain only accepts amino-json proofs, the sign doc being an
			// ADR-036 amino-json doc.
			sig, pub, err := clientCtx.Keyring.Sign(clientCtx.FromName, signBytes, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
			if err != nil {
				return err
			}

			bz, err := clientCtx.TxConfig.MarshalSignatureJSON([]signing.SignatureV2{{
				PubKey: pub,
				Data: &signing.SingleSignatureData{
					SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					Signature: sig,
				},
			}})
			if err != nil {
				return err
			}

			closeFunc, err := setOutputFile(cmd)
			if err != nil {
				return err
			}
			defer closeFunc()

			cmd.Printf("%s\n", bz)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagOutputDocument, "", "The document is written to the given file instead of STDOUT")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewChangePubKeyCmd returns a command that sends a MsgChangePubKey.
func NewChangePubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-pubkey [account-address] [new-pubkey-json]",
		Short: "Replace an account's pubkey while keeping its address",
		Long: `Replace the pubkey of [account-address] with [new-pubkey-json]. The tx is
signed by the account's current key (--from). If the current key's address is not
the account address, because the account was rekeyed before, also pass
--signer-address [account-address].

--proof takes the files written by sign-rekey-proof, comma separated. For a
multisig new pubkey, pass enough member proofs to meet its threshold; they are
combined into one multisignature.

Make sure you control the new pubkey: the account, including staked funds and any
validator it operates, is only reachable through it afterwards.`,
		Example: fmt.Sprintf(`%s tx auth change-pubkey cosmos1... "$(cat new_pubkey.json)" --proof proof1.json,proof2.json --from current-key`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address, newPk, err := parseRekeyArgs(clientCtx, args)
			if err != nil {
				return err
			}

			from, err := clientCtx.TxConfig.SigningContext().AddressCodec().BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			if from != address {
				return fmt.Errorf("account %s is not the account %s the --from key signs for; pass --%s %s if the --from key controls a rekeyed account", address, from, flags.FlagSignerAddress, address)
			}

			proofFiles, err := cmd.Flags().GetString(flagProof)
			if err != nil {
				return err
			}
			proof, err := buildRekeyProof(clientCtx, newPk, proofFiles)
			if err != nil {
				return err
			}

			anyPk, err := codectypes.NewAnyWithValue(newPk)
			if err != nil {
				return err
			}
			msg := &types.MsgChangePubKey{
				Address:   address,
				NewPubKey: anyPk,
				Proof:     proof,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagProof, "", "Comma-separated files written by sign-rekey-proof")
	_ = cmd.MarkFlagRequired(flagProof)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// parseRekeyArgs parses the [account-address] and [new-pubkey-json] args. It
// returns the address in canonical form, since the proof doc and the msg must
// carry the same string.
func parseRekeyArgs(clientCtx client.Context, args []string) (string, cryptotypes.PubKey, error) {
	ac := clientCtx.TxConfig.SigningContext().AddressCodec()
	addrBz, err := ac.StringToBytes(args[0])
	if err != nil {
		return "", nil, fmt.Errorf("invalid account address %s: %w", args[0], err)
	}
	address, err := ac.BytesToString(addrBz)
	if err != nil {
		return "", nil, err
	}

	var newPk cryptotypes.PubKey
	if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[1]), &newPk); err != nil {
		return "", nil, fmt.Errorf("invalid new pubkey: %w", err)
	}
	// Check the key's type and shape before anything calls Address() on it.
	// The tx sig limit is the chain's to enforce.
	if err := types.ValidateRekeyPubKey(nil, newPk, types.Params{TxSigLimit: math.MaxUint64}); err != nil {
		return "", nil, err
	}

	return address, newPk, nil
}

// isRekeyProofSigner reports whether signer can sign a proof for newPk: it is
// newPk itself or a direct member of newPk. Members of nested multisigs are
// not supported, since multisig.AddSignatureV2 only places direct members.
func isRekeyProofSigner(newPk, signer cryptotypes.PubKey) bool {
	msig, ok := newPk.(*kmultisig.LegacyAminoPubKey)
	if !ok {
		return newPk.Equals(signer)
	}
	for _, pk := range msig.GetPubKeys() {
		if pk.Equals(signer) {
			return true
		}
	}
	return false
}

// rekeyAccountNumber returns --account-number if set, else queries it.
func rekeyAccountNumber(cmd *cobra.Command, clientCtx client.Context, address string) (uint64, error) {
	if cmd.Flags().Changed(flags.FlagAccountNumber) {
		return cmd.Flags().GetUint64(flags.FlagAccountNumber)
	}
	if clientCtx.Offline {
		return 0, fmt.Errorf("--%s must be set in offline mode", flags.FlagAccountNumber)
	}
	addr, err := clientCtx.TxConfig.SigningContext().AddressCodec().StringToBytes(address)
	if err != nil {
		return 0, err
	}
	num, _, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, addr)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// buildRekeyProof reads the sign-rekey-proof outputs in files and returns the
// encoded MsgChangePubKey proof. For a multisig newPk the signatures are
// combined into a MultiSignatureData.
func buildRekeyProof(clientCtx client.Context, newPk cryptotypes.PubKey, files string) ([]byte, error) {
	var sigs []signing.SignatureV2
	for _, f := range strings.Split(files, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		bz, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		fileSigs, err := clientCtx.TxConfig.UnmarshalSignatureJSON(bz)
		if err != nil {
			return nil, fmt.Errorf("invalid proof file %s: %w", f, err)
		}
		sigs = append(sigs, fileSigs...)
	}
	if len(sigs) == 0 {
		return nil, fmt.Errorf("--%s has no signatures", flagProof)
	}

	var data signing.SignatureData
	if msig, ok := newPk.(*kmultisig.LegacyAminoPubKey); ok {
		multiData := multisig.NewMultisig(len(msig.PubKeys))
		for _, sig := range sigs {
			if err := multisig.AddSignatureV2(multiData, sig, msig.GetPubKeys()); err != nil {
				return nil, err
			}
		}
		// Each slot is filled at most once, so this counts the signers.
		if n := len(multiData.Signatures); uint32(n) < msig.Threshold {
			return nil, fmt.Errorf("the proof has %d signatures, fewer than the new pubkey's threshold %d", n, msig.Threshold)
		}
		data = multiData
	} else {
		if len(sigs) != 1 {
			return nil, fmt.Errorf("a single-key pubkey needs exactly one proof signature, got %d", len(sigs))
		}
		if !newPk.Equals(sigs[0].PubKey) {
			return nil, errors.New("the proof is not signed by the new pubkey")
		}
		data = sigs[0].Data
	}

	return signing.SignatureDataToProto(data).Marshal()
}
