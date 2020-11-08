package services

import (
	"context"
	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

// interface implementation assertion
var _ crg.ConstructionAPI = SingleNetwork{}

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
	if len(request.Options) == 0 {
		return nil, rosetta.ErrInterpreting.RosettaError()
	}

	addr, ok := request.Options[rosetta.OptionAddress]
	if !ok {
		return nil, rosetta.ErrInvalidAddress.RosettaError()
	}
	addrString := addr.(string)
	accountInfo, err := sn.client.AccountInfo(ctx, addrString, nil)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	gas, ok := request.Options[rosetta.OptionGas]
	if !ok {
		return nil, rosetta.WrapError(rosetta.ErrInvalidAddress, "gas not set").RosettaError()
	}

	memo, ok := request.Options[rosetta.OptionMemo]
	if !ok {
		return nil, rosetta.WrapError(rosetta.ErrInvalidMemo, "memo not set").RosettaError()
	}

	status, err := sn.client.Status(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	res := &types.ConstructionMetadataResponse{
		Metadata: map[string]interface{}{
			rosetta.AccountNumber: accountInfo.GetAccountNumber(),
			rosetta.Sequence:      accountInfo.GetSequence(),
			rosetta.ChainId:       status.NodeInfo.Network,
			rosetta.OptionGas:     gas,
			rosetta.OptionMemo:    memo,
		},
	}

	return res, nil
}

func (sn SingleNetwork) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionPayloads(ctx context.Context, request *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}

func (sn SingleNetwork) ConstructionPreprocess(ctx context.Context, request *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	operations := request.Operations
	if len(operations) != 2 {
		return nil, rosetta.ErrInterpreting.RosettaError()
	}

	txData, err := conversion.GetTransferTxDataFromOperations(operations)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidAddress, err.Error()).RosettaError()
	}
	if txData.FromAddress == "" {
		return nil, rosetta.WrapError(rosetta.ErrInvalidAddress, err.Error()).RosettaError()
	}

	memo, ok := request.Metadata["memo"]
	if !ok {
		return nil, rosetta.ErrInvalidMemo.RosettaError()

	}

	var res = &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			rosetta.OptionAddress: txData.FromAddress,
			rosetta.OptionMemo:    memo,
			rosetta.OptionGas:     request.SuggestedFeeMultiplier,
		},
	}
	return res, nil
}

func (sn SingleNetwork) ConstructionSubmit(ctx context.Context, request *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, rosetta.ErrNotImplemented.RosettaError()
}
