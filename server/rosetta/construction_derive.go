package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (l launchpad) ConstructionDerive(ctx context.Context, r *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	if r.PublicKey.CurveType != "secp256k1" {
		return nil, ErrUnsupportedCurve
	}
	//
	//pubKey, err := secp256k1.ParsePubKey(r.PublicKey.Bytes, secp256k1.S256())
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidPubkey, err.Error())
	//}
	//
	//var pubkeyBytes secp256k12.PubKeySecp256k1
	//copy(pubkeyBytes[:], pubKey.SerializeCompressed())

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			//Address: sdk.AccAddress(pubkeyBytes.Address().Bytes()).String(),
			Address: sdk.AccAddress("").String(),
		},
	}, nil
}
