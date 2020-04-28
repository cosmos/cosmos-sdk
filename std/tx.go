package std

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"

	clientx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var (
	_ sdk.Tx                  = (*Transaction)(nil)
	_ clientx.ClientTx        = (*Transaction)(nil)
	_ clientx.Generator       = TxGenerator{}
	_ clientx.ClientFee       = &StdFee{}
	_ clientx.ClientSignature = &StdSignature{}
)

// TxGenerator defines a transaction generator that allows clients to construct
// transactions.
type TxGenerator struct{}

func (g TxGenerator) NewFee() clientx.ClientFee {
	return &StdFee{}
}

func (g TxGenerator) NewSignature() clientx.ClientSignature {
	return &StdSignature{}
}

// NewTx returns a reference to an empty Transaction type.
func (TxGenerator) NewTx() clientx.ClientTx {
	return &Transaction{}
}

func NewTransaction(fee StdFee, memo string, sdkMsgs []sdk.Msg) (*Transaction, error) {
	tx := &Transaction{
		StdTxBase: NewStdTxBase(fee, nil, memo),
	}

	if err := tx.SetMsgs(sdkMsgs...); err != nil {
		return nil, err
	}

	return tx, nil
}

// GetMsgs returns all the messages in a Transaction as a slice of sdk.Msg.
func (tx Transaction) GetMsgs() []sdk.Msg {
	msgs := make([]sdk.Msg, len(tx.Msgs))

	for i, m := range tx.Msgs {
		msgs[i] = m.GetMsg()
	}

	return msgs
}

// GetSigners returns the addresses that must sign the transaction. Addresses are
// returned in a deterministic order. They are accumulated from the GetSigners
// method for each Msg in the order they appear in tx.GetMsgs(). Duplicate addresses
// will be omitted.
func (tx Transaction) GetSigners() []sdk.AccAddress {
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

// ValidateBasic does a simple and lightweight validation check that doesn't
// require access to any other information.
func (tx Transaction) ValidateBasic() error {
	stdSigs := tx.GetSignatures()

	if tx.Fee.Gas > auth.MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid gas supplied; %d > %d", tx.Fee.Gas, auth.MaxGasWanted,
		)
	}
	if tx.Fee.Amount.IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee, "invalid fee provided: %s", tx.Fee.Amount,
		)
	}
	if len(stdSigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}
	if len(stdSigs) != len(tx.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized, "wrong number of signers; expected %d, got %d", tx.GetSigners(), len(stdSigs),
		)
	}

	return nil
}

// SetMsgs sets the messages for a Transaction. It will overwrite any existing
// messages set.
func (tx *Transaction) SetMsgs(sdkMsgs ...sdk.Msg) error {
	msgs := make([]Message, len(sdkMsgs))
	for i, msg := range sdkMsgs {
		m := &Message{}
		if err := m.SetMsg(msg); err != nil {
			return err
		}

		msgs[i] = *m
	}

	tx.Msgs = msgs
	return nil
}

// GetSignatures returns all the transaction's signatures.
func (tx Transaction) GetSignatures() []sdk.Signature {
	sdkSigs := make([]sdk.Signature, len(tx.Signatures))
	for i, sig := range tx.Signatures {
		sdkSigs[i] = sig
	}

	return sdkSigs
}

// SetSignatures sets the transaction's signatures. It will overwrite any
// existing signatures set.
func (tx *Transaction) SetSignatures(sdkSigs ...clientx.ClientSignature) error {
	sigs := make([]StdSignature, len(sdkSigs))
	for i, sig := range sdkSigs {
		if sig != nil {
			sigs[i] = NewStdSignature(sig.GetPubKey(), sig.GetSignature())
		}
	}

	tx.Signatures = sigs
	return nil
}

// GetFee returns the transaction's fee.
func (tx Transaction) GetFee() sdk.Fee {
	return tx.Fee
}

// SetFee sets the transaction's fee. It will overwrite any existing fee set.
func (tx *Transaction) SetFee(fee clientx.ClientFee) error {
	tx.Fee = NewStdFee(fee.GetGas(), fee.GetAmount())
	return nil
}

// GetMemo returns the transaction's memo.
func (tx Transaction) GetMemo() string {
	return tx.Memo
}

// SetMemo sets the transaction's memo. It will overwrite any existing memo set.
func (tx *Transaction) SetMemo(memo string) {
	tx.Memo = memo
}

// CanonicalSignBytes returns the canonical JSON bytes to sign over for the
// Transaction given a chain ID, account sequence and account number. The JSON
// encoding ensures all field names adhere to their proto definition, default
// values are omitted, and follows the JSON Canonical Form.
func (tx Transaction) CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error) {
	return NewSignDoc(num, seq, cid, tx.Memo, tx.Fee, tx.Msgs...).CanonicalSignBytes()
}

func NewSignDoc(num, seq uint64, cid, memo string, fee StdFee, msgs ...Message) *SignDoc {
	return &SignDoc{
		StdSignDocBase: NewStdSignDocBase(num, seq, cid, memo, fee),
		Msgs:           msgs,
	}
}

// CanonicalSignBytes returns the canonical JSON bytes to sign over, where the
// SignDoc is derived from a Transaction. The JSON encoding ensures all field
// names adhere to their proto definition, default values are omitted, and follows
// the JSON Canonical Form.
func (sd *SignDoc) CanonicalSignBytes() ([]byte, error) {
	return sdk.CanonicalSignBytes(sd)
}

// NewStdFee returns a new instance of StdFee
func NewStdFee(gas uint64, amount sdk.Coins) StdFee {
	return StdFee{
		Amount: amount,
		Gas:    gas,
	}
}

func NewStdSignature(pk crypto.PubKey, sig []byte) StdSignature {
	var pkBz []byte
	if pk != nil {
		pkBz = pk.Bytes()
	}

	return StdSignature{PubKey: pkBz, Signature: sig}
}

func NewStdTxBase(fee StdFee, sigs []StdSignature, memo string) StdTxBase {
	return StdTxBase{
		Fee:        fee,
		Signatures: sigs,
		Memo:       memo,
	}
}

func NewStdSignDocBase(num, seq uint64, cid, memo string, fee StdFee) StdSignDocBase {
	return StdSignDocBase{
		ChainID:       cid,
		AccountNumber: num,
		Sequence:      seq,
		Memo:          memo,
		Fee:           fee,
	}
}

func (m StdFee) GetGas() uint64 {
	return m.Gas
}

func (m StdFee) GetAmount() sdk.Coins {
	return m.Amount
}

func (m *StdFee) SetGas(gas uint64) {
	m.Gas = gas
}

func (m *StdFee) SetAmount(amount sdk.Coins) {
	m.Amount = amount
}

func (m StdSignature) GetPubKey() crypto.PubKey {
	var pk crypto.PubKey
	if len(m.PubKey) == 0 {
		return nil
	}

	amino.MustUnmarshalBinaryBare(m.PubKey, &pk)
	return pk
}

func (m StdSignature) GetSignature() []byte {
	return m.Signature
}

func (m *StdSignature) SetPubKey(pk crypto.PubKey) error {
	m.PubKey = pk.Bytes()
	return nil
}

func (m *StdSignature) SetSignature(signature []byte) {
	m.Signature = signature
}
