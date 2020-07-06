package tx

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultTxDecoder returns a default protobuf TxDecoder using the provided Marshaler and PublicKeyCodec
func DefaultTxDecoder(cdc codec.Marshaler, keyCodec cryptotypes.PublicKeyCodec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var raw tx.TxRaw
		err := cdc.UnmarshalBinaryBare(txBytes, &raw)
		if err != nil {
			return nil, err
		}

		var theTx tx.Tx
		err = cdc.UnmarshalBinaryBare(txBytes, &theTx)
		if err != nil {
			return nil, err
		}

		pks, err := extractPubKeys(theTx, keyCodec)
		if err != nil {
			return nil, err
		}

		return &builder{
			tx:          &theTx,
			bodyBz:      raw.BodyBytes,
			authInfoBz:  raw.AuthInfoBytes,
			pubKeys:     pks,
			marshaler:   cdc,
			pubkeyCodec: keyCodec,
		}, nil
	}
}

// DefaultTxDecoder returns a default protobuf JSON TxDecoder using the provided Marshaler and PublicKeyCodec
func DefaultJSONTxDecoder(cdc codec.Marshaler, keyCodec cryptotypes.PublicKeyCodec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var theTx tx.Tx
		err := cdc.UnmarshalJSON(txBytes, &theTx)
		if err != nil {
			return nil, err
		}

		pks, err := extractPubKeys(theTx, keyCodec)
		if err != nil {
			return nil, err
		}

		return &builder{
			tx:          &theTx,
			pubKeys:     pks,
			marshaler:   cdc,
			pubkeyCodec: keyCodec,
		}, nil
	}
}

func extractPubKeys(tx tx.Tx, keyCodec cryptotypes.PublicKeyCodec) ([]crypto.PubKey, error) {
	if tx.AuthInfo == nil {
		return []crypto.PubKey{}, nil
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
	return pks, nil
}
