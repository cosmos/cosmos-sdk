package signing

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type DecodedTx struct {
	*types.Tx
	Raw *TxRaw
}

var _ types.ProtoTx = DecodedTx{}

func DefaultTxDecoder(cdc codec.Marshaler, keyCodec cryptotypes.PublicKeyCodec) sdk.TxDecoder {
	keyCodec = cryptotypes.CacheWrapCodec(keyCodec)

	return func(txBytes []byte) (sdk.Tx, error) {
		var raw TxRaw
		err := cdc.UnmarshalBinaryBare(txBytes, &raw)
		if err != nil {
			return nil, err
		}

		var tx types.Tx
		err = cdc.UnmarshalBinaryBare(txBytes, &tx)
		if err != nil {
			return nil, err
		}

		return DecodedTx{
			Tx:  &tx,
			Raw: &raw,
		}, nil
	}
}

func (m DecodedTx) GetBodyBytes() []byte {
	return m.Raw.BodyBytes
}

func (m DecodedTx) GetAuthInfoBytes() []byte {
	return m.Raw.AuthInfoBytes
}
