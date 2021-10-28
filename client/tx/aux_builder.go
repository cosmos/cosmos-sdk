package tx

import (
	"github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// AuxTxBuilder is a client-side builder for creating an AuxTx.
type AuxTxBuilder struct {
	body  *tx.TxBody
	auxTx *tx.AuxTx
}

func (b *AuxTxBuilder) SetMemo(memo string) {
	b.checkEmptyFields()

	b.body.Memo = memo
}

func (b *AuxTxBuilder) SetTimeoutHeight(height uint64) {
	b.checkEmptyFields()

	b.body.TimeoutHeight = height
}

func (b *AuxTxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	anys := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
	}

	b.checkEmptyFields()

	b.body.Messages = anys

	return nil
}

func (b *AuxTxBuilder) SetAccountNumber(accNum uint64) {
	b.checkEmptyFields()

	b.auxTx.SignDoc.AccountNumber = accNum
}

func (b *AuxTxBuilder) SetChainID(chainID string) {
	b.checkEmptyFields()

	b.auxTx.SignDoc.ChainId = chainID
}

func (b *AuxTxBuilder) SetSequence(accSeq uint64) {
	b.checkEmptyFields()

	b.auxTx.SignDoc.Sequence = accSeq
}

func (b *AuxTxBuilder) SetPubKey(pk cryptotypes.PubKey) error {
	b.checkEmptyFields()

	any, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return err
	}

	b.auxTx.SignDoc.PublicKey = any

	return nil
}

func (b *AuxTxBuilder) SetSignature(sig []byte) {
	if b.auxTx == nil {
		b.auxTx = &tx.AuxTx{}
	}

	b.auxTx.Sig = sig
}

// GetSignBytes returns the builder's sign bytes.
func (b *AuxTxBuilder) GetSignBytes() ([]byte, error) {
	body := b.body
	if body == nil {
		return nil, sdkerrors.ErrLogic.Wrap("tx body is nil, call setters on AuxTxBuilder first")
	}

	bodyBz, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}

	auxTx := b.auxTx
	if auxTx == nil {
		return nil, sdkerrors.ErrLogic.Wrap("aux tx is nil, call setters on AuxTxBuilder first")
	}

	sd := auxTx.SignDoc
	if sd == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign doc is nil, call setters on AuxTxBuilder first")
	}

	sd.BodyBytes = bodyBz

	if err := b.auxTx.SignDoc.ValidateBasic(); err != nil {
		return nil, err
	}

	signBz, err := proto.Marshal(b.auxTx.SignDoc)
	if err != nil {
		return nil, err
	}

	return signBz, nil
}

// GetAuxTx returns the builder's AuxTx.
func (b *AuxTxBuilder) GetAuxTx() (*tx.AuxTx, error) {
	if err := b.auxTx.ValidateBasic(); err != nil {
		return nil, err
	}

	return b.auxTx, nil
}

func (b *AuxTxBuilder) checkEmptyFields() {
	if b.body == nil {
		b.body = &tx.TxBody{}
	}

	if b.auxTx == nil {
		b.auxTx = &tx.AuxTx{}
		if b.auxTx.SignDoc == nil {
			b.auxTx.SignDoc = &tx.SignDocDirectAux{}
		}
	}
}
