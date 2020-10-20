package rosetta

import (
	"context"

	secp256k12 "github.com/tendermint/tendermint/crypto/secp256k1"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (l launchpad) ConstructionDerive(ctx context.Context, r *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	if r.PublicKey.CurveType != "secp256k1" {
		return nil, ErrUnsupportedCurve
	}

	pubKey, err := secp256k1.ParsePubKey(r.PublicKey.Bytes, secp256k1.S256())
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidPubkey, err.Error())
	}

	var pubkeyBytes secp256k12.PubKeySecp256k1
	copy(pubkeyBytes[:], pubKey.SerializeCompressed())

	return &types.ConstructionDeriveResponse{
		Address: sdk.AccAddress(pubkeyBytes.Address().Bytes()).String(),
	}, nil
}
