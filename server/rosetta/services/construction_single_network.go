package services

import (
	"context"
	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (sn SingleNetwork) ConstructionCombine(ctx context.Context, request *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionDerive(ctx context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	if request.PublicKey.CurveType != "secp256k1" {
		return nil, rosetta.WrapError(rosetta.ErrUnsupportedCurve, "only secp256k1 supported").RosettaError()
	}

	cmp, err := btcec.ParsePubKey(request.PublicKey.Bytes, btcec.S256())
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	compressedPublicKey := make([]byte, secp256k1.PubKeySize)
	copy(compressedPublicKey, cmp.SerializeCompressed())

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: sdk.AccAddress(compressedPublicKey).String(),
		},
	}, nil
}

func (sn SingleNetwork) ConstructionHash(ctx context.Context, request *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionMetadata(ctx context.Context, request *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionPayloads(ctx context.Context, request *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionPreprocess(ctx context.Context, request *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionSubmit(ctx context.Context, request *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}
