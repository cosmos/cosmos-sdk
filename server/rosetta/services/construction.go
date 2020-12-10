package services

import (
	"context"
	"encoding/hex"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
)

func (on OnlineNetwork) ConstructionCombine(ctx context.Context, request *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.UnsignedTransaction)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidTransaction, "error decoding unsigned tx").RosettaError()
	}

	signedTxBytes, err := on.client.SignedTx(ctx, txBytes, request.Signatures)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.ConstructionCombineResponse{SignedTransaction: hex.EncodeToString(signedTxBytes)}, nil
}

func (on OnlineNetwork) ConstructionDerive(_ context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	account, err := on.client.AccountIdentifierFromPubKeyBytes((string)(request.PublicKey.CurveType), request.PublicKey.Bytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: account,
		Metadata:          nil,
	}, nil
}

func (on OnlineNetwork) ConstructionHash(_ context.Context, request *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		return nil, rosetta.ErrInvalidTransaction.RosettaError()
	}
	txIdentifier, err := on.client.TransactionIdentifierFromHexBytes(txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: txIdentifier,
		Metadata:              nil,
	}, nil
}

func (on OnlineNetwork) ConstructionMetadata(ctx context.Context, request *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	meta, err := on.client.ConstructionMetadataFromOptions(ctx, request.Options)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return &types.ConstructionMetadataResponse{
		Metadata:     meta,
		SuggestedFee: nil,
	}, nil
}

func (on OnlineNetwork) ConstructionParse(_ context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.Transaction)
	if err != nil {
		return nil, rosetta.ErrInvalidTransaction.RosettaError()
	}
	ops, signers, err := on.client.TxOperationsAndSignersAccountIdentifiers(request.Signed, txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return &types.ConstructionParseResponse{
		Operations:               ops,
		AccountIdentifierSigners: signers,
		Metadata:                 nil,
	}, nil
}

func (on OnlineNetwork) ConstructionPayloads(ctx context.Context, request *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	payload, err := on.client.ConstructionPayload(ctx, request)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return payload, nil
}

func (on OnlineNetwork) ConstructionPreprocess(ctx context.Context, request *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	if len(request.Operations) == 0 {
		return nil, rosetta.WrapError(rosetta.ErrBadArgument, "no operations").RosettaError()
	}

	metadata, err := on.client.OperationsToMetadata(ctx, request.Operations)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	memo, ok := request.Metadata["memo"]
	if !ok {
		memo = ""
	}
	metadata["memo"] = memo

	return &types.ConstructionPreprocessResponse{
		Options: metadata,
	}, nil
}

func (on OnlineNetwork) ConstructionSubmit(ctx context.Context, request *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		return nil, rosetta.ErrInvalidTransaction.RosettaError()
	}

	txIdentifier, meta, err := on.client.PostTxBytes(ctx, txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: txIdentifier,
		Metadata:              meta,
	}, nil
}
