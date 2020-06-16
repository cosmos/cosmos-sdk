package tx

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultTxDecoder(cdc codec.Marshaler, keyCodec cryptotypes.PublicKeyCodec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var raw TxRaw
		err := cdc.UnmarshalBinaryBare(txBytes, &raw)
		if err != nil {
			return nil, err
		}

		var tx Tx
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

		return builder{
			tx:          &tx,
			bodyBz:      raw.BodyBytes,
			authInfoBz:  raw.AuthInfoBytes,
			pubKeys:     pks,
			marshaler:   cdc,
			pubkeyCodec: keyCodec,
		}, nil
	}
}
