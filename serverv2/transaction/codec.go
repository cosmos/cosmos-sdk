package transaction

import (
	"fmt"

	v1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	txdecoder "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

type Codec struct {
	decoder *txdecoder.Decoder
}

func NewCodec() Codec {
	return Codec{}
}

func (c Codec) RegisterCodec(sc *signing.Context) error {
	decoder, err := txdecoder.NewDecoder(txdecoder.Options{SigningContext: sc})
	if err != nil {
		return err
	}
	c.decoder = decoder
	return nil
}

// DefaultTxEncoder returns a default protobuf TxEncoder using the provided Marshaler
func Encode(tx sdk.Tx) ([]byte, error) {
	txWrapper, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", &wrapper{}, tx)
	}

	raw := &txtypes.TxRaw{
		BodyBytes:     txWrapper.getBodyBytes(),
		AuthInfoBytes: txWrapper.getAuthInfoBytes(),
		Signatures:    txWrapper.tx.Signatures,
	}

	return proto.Marshal(raw)
}

func (c Codec) Decode(txBytes []byte) (v1beta1.Tx, error) {

	tx, err := c.decoder.Decode(txBytes)

	return tx, err
}

// wrapper is a wrapper around the tx.Tx proto.Message which retain the raw
// body and auth_info bytes.
type wrapper struct {
	cdc codec.Codec

	tx *txtypes.Tx

	// bodyBz represents the protobuf encoding of TxBody. This should be encoding
	// from the client using TxRaw if the tx was decoded from the wire
	bodyBz []byte

	// authInfoBz represents the protobuf encoding of TxBody. This should be encoding
	// from the client using TxRaw if the tx was decoded from the wire
	authInfoBz []byte

	txBodyHasUnknownNonCriticals bool

	signers [][]byte
	msgsV2  []protov2.Message
}

func (w *wrapper) getBodyBytes() []byte {
	if len(w.bodyBz) == 0 {
		// if bodyBz is empty, then marshal the body. bodyBz will generally
		// be set to nil whenever SetBody is called so the result of calling
		// this method should always return the correct bytes. Note that after
		// decoding bodyBz is derived from TxRaw so that it matches what was
		// transmitted over the wire
		var err error
		w.bodyBz, err = proto.Marshal(w.tx.Body)
		if err != nil {
			panic(err)
		}
	}
	return w.bodyBz
}

func (w *wrapper) getAuthInfoBytes() []byte {
	if len(w.authInfoBz) == 0 {
		// if authInfoBz is empty, then marshal the body. authInfoBz will generally
		// be set to nil whenever SetAuthInfo is called so the result of calling
		// this method should always return the correct bytes. Note that after
		// decoding authInfoBz is derived from TxRaw so that it matches what was
		// transmitted over the wire
		var err error
		w.authInfoBz, err = proto.Marshal(w.tx.AuthInfo)
		if err != nil {
			panic(err)
		}
	}
	return w.authInfoBz
}

func (w *wrapper) GetMsgs() []sdk.Msg {
	return w.tx.GetMsgs()
}

func (w *wrapper) GetMsgsV2() ([]protov2.Message, error) {
	if w.msgsV2 == nil {
		err := w.initSignersAndMsgsV2()
		if err != nil {
			return nil, err
		}
	}

	return w.msgsV2, nil
}

func (w *wrapper) initSignersAndMsgsV2() error {
	var err error
	w.signers, w.msgsV2, err = w.tx.GetSigners(w.cdc)
	return err
}

func (w *wrapper) GetSigners() ([][]byte, error) {
	if w.signers == nil {
		err := w.initSignersAndMsgsV2()
		if err != nil {
			return nil, err
		}
	}
	return w.signers, nil
}
