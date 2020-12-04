package rosetta

import (
	"context"
	"encoding/hex"
	"fmt"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	secp256k12 "github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

// ConstructionCombine implements the /construction/combine endpoint.
func (l launchpad) ConstructionCombine(ctx context.Context, r *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	bz, err := hex.DecodeString(r.UnsignedTransaction)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "error decoding unsigned tx")
	}

	var stdTx auth.StdTx
	err = l.cdc.UnmarshalJSON(bz, &stdTx)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, fmt.Sprintf("unable to unmarshal tx: %s", err.Error()))
	}

	sigs := make([]auth.StdSignature, len(r.Signatures))
	for i, signature := range r.Signatures {
		if signature.PublicKey.CurveType != "secp256k1" {
			return nil, ErrUnsupportedCurve
		}

		pubKey, err := secp256k1.ParsePubKey(signature.PublicKey.Bytes, secp256k1.S256())
		if err != nil {
			return nil, rosetta.WrapError(ErrInvalidPubkey, err.Error())
		}

		var compressedPublicKey secp256k12.PubKeySecp256k1
		copy(compressedPublicKey[:], pubKey.SerializeCompressed())

		sign := auth.StdSignature{
			PubKey:    compressedPublicKey,
			Signature: signature.Bytes,
		}
		sigs[i] = sign
	}

	stdTx.Signatures = sigs
	txBytes, err := l.cdc.MarshalJSON(stdTx)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "unable to marshal signed tx")
	}
	txHex := hex.EncodeToString(txBytes)

	return &types.ConstructionCombineResponse{
		SignedTransaction: txHex,
	}, nil
}
