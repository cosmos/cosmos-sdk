package signing

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type DecodedTx struct {
	*types.Tx
	Raw     *TxRaw
	PubKeys []crypto.PubKey
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

		signerInfos := tx.AuthInfo.SignerInfos
		pks := make([]crypto.PubKey, len(signerInfos))
		for i, si := range signerInfos {
			pk, err := keyCodec.Decode(si.PublicKey)
			if err != nil {
				return nil, err
			}
			pks[i] = pk
		}

		return DecodedTx{
			Tx:      &tx,
			Raw:     &raw,
			PubKeys: pks,
		}, nil
	}
}

func (m DecodedTx) GetBodyBytes() []byte {
	return m.Raw.BodyBytes
}

func (m DecodedTx) GetAuthInfoBytes() []byte {
	return m.Raw.AuthInfoBytes
}

func (m DecodedTx) GetPubKeys() []crypto.PubKey {
	return m.PubKeys
}
