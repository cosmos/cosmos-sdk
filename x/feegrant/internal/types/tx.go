package types

import (
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ sdk.Tx = FeeGrantTx{}

	maxGasWanted = uint64((1 << 63) - 1)
)

// FeeGrantTx wraps a Msg with Fee and Signatures,
// adding the ability to delegate the fee payment
// NOTE: the first signature responsible for paying fees, either directly,
// or must be authorized to spend from the provided Fee.FeeAccount
type FeeGrantTx struct {
	Msgs       []sdk.Msg                `json:"msg" yaml:"msg"`
	Fee        GrantedFee               `json:"fee" yaml:"fee"`
	Signatures []authtypes.StdSignature `json:"signatures" yaml:"signatures"`
	Memo       string                   `json:"memo" yaml:"memo"`
	FeeAccount sdk.AccAddress           `json:"fee_account" yaml:"fee_account"`
}

func NewFeeGrantTx(msgs []sdk.Msg, fee GrantedFee, sigs []authtypes.StdSignature, memo string) FeeGrantTx {
	return FeeGrantTx{
		Msgs:       msgs,
		Fee:        fee,
		Signatures: sigs,
		Memo:       memo,
	}
}

// GetMsgs returns the all the transaction's messages.
func (tx FeeGrantTx) GetMsgs() []sdk.Msg { return tx.Msgs }

// ValidateBasic does a simple and lightweight validation check that doesn't
// require access to any other information.
func (tx FeeGrantTx) ValidateBasic() sdk.Error {
	stdSigs := tx.GetSignatures()

	if tx.Fee.Gas > maxGasWanted {
		return sdk.ErrGasOverflow(fmt.Sprintf("invalid gas supplied; %d > %d", tx.Fee.Gas, maxGasWanted))
	}
	if tx.Fee.Amount.IsAnyNegative() {
		return sdk.ErrInsufficientFee(fmt.Sprintf("invalid fee %s amount provided", tx.Fee.Amount))
	}
	if len(stdSigs) == 0 {
		return sdk.ErrNoSignatures("no signers")
	}
	if len(stdSigs) != len(tx.GetSigners()) {
		return sdk.ErrUnauthorized("wrong number of signers")
	}

	return nil
}

// CountSubKeys counts the total number of keys for a multi-sig public key.
func CountSubKeys(pub crypto.PubKey) int {
	v, ok := pub.(multisig.PubKeyMultisigThreshold)
	if !ok {
		return 1
	}

	numKeys := 0
	for _, subkey := range v.PubKeys {
		numKeys += CountSubKeys(subkey)
	}

	return numKeys
}

// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// They are accumulated from the GetSigners method for each Msg
// in the order they appear in tx.GetMsgs().
// Duplicate addresses will be omitted.
func (tx FeeGrantTx) GetSigners() []sdk.AccAddress {
	seen := map[string]bool{}
	var signers []sdk.AccAddress
	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}
	return signers
}

// GetMemo returns the memo
func (tx FeeGrantTx) GetMemo() string { return tx.Memo }

// GetSignatures returns the signature of signers who signed the Msg.
// CONTRACT: Length returned is same as length of
// pubkeys returned from MsgKeySigners, and the order
// matches.
// CONTRACT: If the signature is missing (ie the Msg is
// invalid), then the corresponding signature is
// .Empty().
func (tx FeeGrantTx) GetSignatures() [][]byte {
	sigs := make([][]byte, len(tx.Signatures))
	for i, stdSig := range tx.Signatures {
		sigs[i] = stdSig.Signature
	}
	return sigs
}

// GetPubkeys returns the pubkeys of signers if the pubkey is included in the signature
// If pubkey is not included in the signature, then nil is in the slice instead
func (tx FeeGrantTx) GetPubKeys() []crypto.PubKey {
	pks := make([]crypto.PubKey, len(tx.Signatures))
	for i, stdSig := range tx.Signatures {
		pks[i] = stdSig.PubKey
	}
	return pks
}

// GetSignBytes returns the signBytes of the tx for a given signer
func (tx FeeGrantTx) GetSignBytes(ctx sdk.Context, acc exported.Account) []byte {
	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}
	return StdSignBytes(chainID, accNum, acc.GetSequence(), tx.Fee, tx.Msgs, tx.Memo)
}

// GetGas returns the Gas in GrantedFee
func (tx FeeGrantTx) GetGas() uint64 { return tx.Fee.Gas }

// GetFee returns the FeeAmount in GrantedFee
func (tx FeeGrantTx) GetFee() sdk.Coins { return tx.Fee.Amount }

// FeePayer returns the address that is responsible for paying fee
// This can be explicily set in GrantedFee, or defaults to MainSigner
func (tx FeeGrantTx) FeePayer() sdk.AccAddress {
	if len(tx.Fee.FeeAccount) != 0 {
		return tx.Fee.FeeAccount
	}
	return tx.MainSigner()
}

// MainSigner returns the first signer of the tx, by default this
// account is responsible for fees, if not explicitly set.
func (tx FeeGrantTx) MainSigner() sdk.AccAddress {
	if len(tx.GetSigners()) != 0 {
		return tx.GetSigners()[0]
	}
	return sdk.AccAddress{}
}

// GrantedFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type GrantedFee struct {
	Amount     sdk.Coins      `json:"amount" yaml:"amount"`
	Gas        uint64         `json:"gas" yaml:"gas"`
	FeeAccount sdk.AccAddress `json:"fee_account,omitempty" yaml:"fee_account"`
}

// NewGrantedFee returns a new instance of GrantedFee
func NewGrantedFee(gas uint64, amount sdk.Coins, feeAccount sdk.AccAddress) GrantedFee {
	return GrantedFee{
		Amount:     amount,
		Gas:        gas,
		FeeAccount: feeAccount,
	}
}

// Bytes for signing later
func (fee GrantedFee) Bytes() []byte {
	// normalize. XXX
	// this is a sign of something ugly
	// (in the lcd_test, client side its null,
	// server side its [])
	if len(fee.Amount) == 0 {
		fee.Amount = sdk.NewCoins()
	}
	cdc := codec.New()
	bz, err := cdc.MarshalJSON(fee)
	if err != nil {
		panic(err)
	}
	return bz
}

// GasPrices returns the gas prices for a GrantedFee.
//
// NOTE: The gas prices returned are not the true gas prices that were
// originally part of the submitted transaction because the fee is computed
// as fee = ceil(gasWanted * gasPrices).
func (fee GrantedFee) GasPrices() sdk.DecCoins {
	return sdk.NewDecCoins(fee.Amount).QuoDec(sdk.NewDec(int64(fee.Gas)))
}

// DelegatedSignDoc is replay-prevention structure.
// It includes the result of msg.GetSignBytes(),
// as well as the ChainID (prevent cross chain replay)
// and the Sequence numbers for each signature (prevent
// inchain replay and enforce tx ordering per account).
type DelegatedSignDoc struct {
	AccountNumber uint64            `json:"account_number" yaml:"account_number"`
	ChainID       string            `json:"chain_id" yaml:"chain_id"`
	Fee           json.RawMessage   `json:"fee" yaml:"fee"`
	Memo          string            `json:"memo" yaml:"memo"`
	Msgs          []json.RawMessage `json:"msgs" yaml:"msgs"`
	Sequence      uint64            `json:"sequence" yaml:"sequence"`
}

// StdSignBytes returns the bytes to sign for a transaction.
func StdSignBytes(chainID string, accnum uint64, sequence uint64, fee GrantedFee, msgs []sdk.Msg, memo string) []byte {
	cdc := codec.New()
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		msgsBytes = append(msgsBytes, json.RawMessage(msg.GetSignBytes()))
	}
	bz, err := cdc.MarshalJSON(DelegatedSignDoc{
		AccountNumber: accnum,
		ChainID:       chainID,
		Fee:           json.RawMessage(fee.Bytes()),
		Memo:          memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}
