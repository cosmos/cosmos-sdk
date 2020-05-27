package signing

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

func DefaultJSONTxDecoder(cdc codec.Marshaler, keyCodec cryptotypes.PublicKeyCodec) sdk.TxDecoder {
	keyCodec = cryptotypes.CacheWrapCodec(keyCodec)

	return func(txBytes []byte) (sdk.Tx, error) {
		var tx types.Tx
		err := cdc.UnmarshalJSON(txBytes, &tx)
		if err != nil {
			return nil, err
		}

		// this decodes pubkeys and makes sure they are cached
		signerInfos := tx.AuthInfo.SignerInfos
		for _, si := range signerInfos {
			_, err := keyCodec.Decode(si.PublicKey)
			if err != nil {
				return nil, err
			}
		}

		return &tx, nil
	}
}
