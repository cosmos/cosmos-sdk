package legacytx

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// Interface implementation checks
var (
	_ sdk.Tx                             = (*StdTx)(nil)
	_ sdk.TxWithMemo                     = (*StdTx)(nil)
	_ sdk.FeeTx                          = (*StdTx)(nil)
	_ codectypes.UnpackInterfacesMessage = (*StdTx)(nil)

	_ codectypes.UnpackInterfacesMessage = (*StdSignature)(nil)
)

// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
// [Deprecated]
type StdFee struct {
	Amount sdk.Coins `json:"amount" yaml:"amount"`
	Gas    uint64    `json:"gas" yaml:"gas"`
}

// Deprecated: NewStdFee returns a new instance of StdFee
func NewStdFee(gas uint64, amount sdk.Coins) StdFee {
	return StdFee{
		Amount: amount,
		Gas:    gas,
	}
}

// GetGas returns the fee's (wanted) gas.
func (fee StdFee) GetGas() uint64 {
	return fee.Gas
}

// GetAmount returns the fee's amount.
func (fee StdFee) GetAmount() sdk.Coins {
	return fee.Amount
}

// Bytes returns the encoded bytes of a StdFee.
func (fee StdFee) Bytes() []byte {
	if len(fee.Amount) == 0 {
		fee.Amount = sdk.NewCoins()
	}

	bz, err := legacy.Cdc.MarshalJSON(fee)
	if err != nil {
		panic(err)
	}

	return bz
}

// GasPrices returns the gas prices for a StdFee.
//
// NOTE: The gas prices returned are not the true gas prices that were
// originally part of the submitted transaction because the fee is computed
// as fee = ceil(gasWanted * gasPrices).
func (fee StdFee) GasPrices() sdk.DecCoins {
	return sdk.NewDecCoinsFromCoins(fee.Amount...).QuoDec(sdk.NewDec(int64(fee.Gas)))
}

// StdTx is the legacy transaction format for wrapping a Msg with Fee and Signatures.
// It only works with Amino, please prefer the new protobuf Tx in types/tx.
// NOTE: the first signature is the fee payer (Signatures must not be nil).
// Deprecated
type StdTx struct {
	Msgs          []sdk.Msg      `json:"msg" yaml:"msg"`
	Fee           StdFee         `json:"fee" yaml:"fee"`
	Signatures    []StdSignature `json:"signatures" yaml:"signatures"`
	Memo          string         `json:"memo" yaml:"memo"`
	TimeoutHeight uint64         `json:"timeout_height" yaml:"timeout_height"`
}

// Deprecated
func NewStdTx(msgs []sdk.Msg, fee StdFee, sigs []StdSignature, memo string) StdTx {
	return StdTx{
		Msgs:       msgs,
		Fee:        fee,
		Signatures: sigs,
		Memo:       memo,
	}
}

// GetMsgs returns the all the transaction's messages.
func (tx StdTx) GetMsgs() []sdk.Msg { return tx.Msgs }

// ValidateBasic does a simple and lightweight validation check that doesn't
// require access to any other information.
func (tx StdTx) ValidateBasic() error {
	stdSigs := tx.GetSignatures()

	if tx.Fee.Gas > txtypes.MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", tx.Fee.Gas, txtypes.MaxGasWanted,
		)
	}
	if tx.Fee.Amount.IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"invalid fee provided: %s", tx.Fee.Amount,
		)
	}
	if len(stdSigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}
	if len(stdSigs) != len(tx.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"wrong number of signers; expected %d, got %d", len(tx.GetSigners()), len(stdSigs),
		)
	}

	return nil
}

// Deprecated: AsAny implements intoAny. It doesn't work for protobuf serialization,
// so it can't be saved into protobuf configured storage. We are using it only for API
// compatibility.
func (tx *StdTx) AsAny() *codectypes.Any {
	return codectypes.UnsafePackAny(tx)
}

// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// They are accumulated from the GetSigners method for each Msg
// in the order they appear in tx.GetMsgs().
// Duplicate addresses will be omitted.
func (tx StdTx) GetSigners() []sdk.AccAddress {
	var signers []sdk.AccAddress
	seen := map[string]bool{}

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
func (tx StdTx) GetMemo() string { return tx.Memo }

// GetTimeoutHeight returns the transaction's timeout height (if set).
func (tx StdTx) GetTimeoutHeight() uint64 {
	return tx.TimeoutHeight
}

// GetSignatures returns the signature of signers who signed the Msg.
// CONTRACT: Length returned is same as length of
// pubkeys returned from MsgKeySigners, and the order
// matches.
// CONTRACT: If the signature is missing (ie the Msg is
// invalid), then the corresponding signature is
// .Empty().
func (tx StdTx) GetSignatures() [][]byte {
	sigs := make([][]byte, len(tx.Signatures))
	for i, stdSig := range tx.Signatures {
		sigs[i] = stdSig.Signature
	}
	return sigs
}

// GetSignaturesV2 implements SigVerifiableTx.GetSignaturesV2
func (tx StdTx) GetSignaturesV2() ([]signing.SignatureV2, error) {
	res := make([]signing.SignatureV2, len(tx.Signatures))

	for i, sig := range tx.Signatures {
		var err error
		res[i], err = StdSignatureToSignatureV2(legacy.Cdc, sig)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "Unable to convert signature %v to V2", sig)
		}
	}

	return res, nil
}

// GetPubkeys returns the pubkeys of signers if the pubkey is included in the signature
// If pubkey is not included in the signature, then nil is in the slice instead
func (tx StdTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	pks := make([]cryptotypes.PubKey, len(tx.Signatures))

	for i, stdSig := range tx.Signatures {
		pks[i] = stdSig.GetPubKey()
	}

	return pks, nil
}

// GetGas returns the Gas in StdFee
func (tx StdTx) GetGas() uint64 { return tx.Fee.Gas }

// GetFee returns the FeeAmount in StdFee
func (tx StdTx) GetFee() sdk.Coins { return tx.Fee.Amount }

// FeePayer returns the address that is responsible for paying fee
// StdTx returns the first signer as the fee payer
// If no signers for tx, return empty address
func (tx StdTx) FeePayer() sdk.AccAddress {
	if tx.GetSigners() != nil {
		return tx.GetSigners()[0]
	}
	return sdk.AccAddress{}
}

// FeeGranter always returns nil for StdTx
func (tx StdTx) FeeGranter() sdk.AccAddress {
	return nil
}

func (tx StdTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, m := range tx.Msgs {
		err := codectypes.UnpackInterfaces(m, unpacker)
		if err != nil {
			return err
		}
	}

	// Signatures contain PubKeys, which need to be unpacked.
	for _, s := range tx.Signatures {
		err := s.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}

	return nil
}
