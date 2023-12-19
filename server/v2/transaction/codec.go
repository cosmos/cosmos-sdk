package transaction

import (
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/transaction"
	txdecoder "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"
)

var _ transaction.Codec[transaction.Tx] = Codec[txdecoder.DecodedTx]{}

// Codec implements the transaction.Codec interface.
// It uses the txdecoder.Decoder to decode and encode transactions.
type Codec[T txdecoder.DecodedTx] struct {
	decoder *txdecoder.Decoder
}

// NewCodec creates a new Codec with usage of the txdecoder.Decoder, located in x/tx
func NewCodec[T txdecoder.DecodedTx]() Codec[T] {
	return Codec[T]{}
}

// RegisterCodec registers the txdecoder.Decoder to the signing context.
func (c *Codec[T]) RegisterCodec(sc *signing.Context) error {
	decoder, err := txdecoder.NewDecoder(txdecoder.Options{SigningContext: sc})
	if err != nil {
		return err
	}
	c.decoder = decoder

	return nil
}

// Decode decodes the transaction bytes into a transaction.Tx.
func (c Codec[T]) Decode(txBytes []byte) (transaction.Tx, error) {
	tx, err := c.decoder.Decode(txBytes)

	return tx, err
}

// Encode encodes the transaction into bytes.
func (c Codec[T]) Encode(tx transaction.Tx) ([]byte, error) {
	dtx, ok := tx.(*txdecoder.DecodedTx)
	if !ok {
		return nil, ErrorInvalidTransaction
	}

	return proto.Marshal(dtx.TxRaw)
}
