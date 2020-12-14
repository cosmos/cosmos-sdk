package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// interface implementation assertion
var _ crg.ConstructionAPI = OnlineNetwork{}

func (on OnlineNetwork) ConstructionCombine(ctx context.Context, request *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.UnsignedTransaction)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	signedTx, err := on.client.SignedTx(ctx, txBytes, request.Signatures)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(signedTx),
	}, nil
}

func (on OnlineNetwork) getTxBuilderFromBytesTx(tx string) (client.TxBuilder, error) {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return nil, err
	}

	TxConfig := on.client.GetTxConfig()
	rawTx, err := TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}

	txBldr, _ := TxConfig.WrapTxBuilder(rawTx)

	return txBldr, nil
}

func (on OnlineNetwork) ConstructionDerive(ctx context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	if request.PublicKey.CurveType != "secp256k1" {
		return nil, rosetta.WrapError(rosetta.ErrUnsupportedCurve, "only secp256k1 supported").RosettaError()
	}

	cmp, err := btcec.ParsePubKey(request.PublicKey.Bytes, btcec.S256())
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	compressedPublicKey := make([]byte, secp256k1.PubKeySize)
	copy(compressedPublicKey, cmp.SerializeCompressed())

	pk := secp256k1.PubKey{Key: compressedPublicKey}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: sdk.AccAddress(pk.Address()).String(),
		},
	}, nil
}

func (on OnlineNetwork) ConstructionHash(ctx context.Context, request *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	bz, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidTransaction, "error decoding tx").RosettaError()
	}

	hash := sha256.Sum256(bz)
	bzHash := hash[:]
	hashString := hex.EncodeToString(bzHash)

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: strings.ToUpper(hashString),
		},
	}, nil
}

func (on OnlineNetwork) ConstructionMetadata(ctx context.Context, request *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	metadata, err := on.client.ConstructionMetadataFromOptions(ctx, request.Options)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.ConstructionMetadataResponse{
		Metadata: metadata,
	}, nil
}

func (on OnlineNetwork) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
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
	if len(request.Operations) > 3 {
		return nil, rosetta.ErrInvalidOperation.RosettaError()
	}

	msgs, signAddr, fee, err := conversion.RosettaOperationsToSdkMsg(request.Operations)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidOperation, err.Error()).RosettaError()
	}

	metadata, err := GetMetadataFromPayloadReq(request)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidRequest, err.Error()).RosettaError()
	}

	txFactory := tx.Factory{}.WithAccountNumber(metadata.AccountNumber).WithChainID(metadata.ChainID).
		WithGas(metadata.Gas).WithSequence(metadata.Sequence).WithMemo(metadata.Memo).WithFees(fee.String())

	TxConfig := on.client.GetTxConfig()
	txFactory = txFactory.WithTxConfig(TxConfig)

	txBldr, err := tx.BuildUnsignedTx(txFactory, msgs...)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	if txFactory.SignMode() == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		txFactory = txFactory.WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	signerData := authsigning.SignerData{
		ChainID:       txFactory.ChainID(),
		AccountNumber: txFactory.AccountNumber(),
		Sequence:      txFactory.Sequence(),
	}

	signBytes, err := TxConfig.SignModeHandler().GetSignBytes(txFactory.SignMode(), signerData, txBldr.GetTx())
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	txBytes, err := TxConfig.TxEncoder()(txBldr.GetTx())
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads: []*types.SigningPayload{
			{
				AccountIdentifier: &types.AccountIdentifier{
					Address: signAddr,
				},
				Bytes:         crypto.Sha256(signBytes),
				SignatureType: "ecdsa",
			},
		},
	}, nil
}

func (on OnlineNetwork) ConstructionPreprocess(ctx context.Context, request *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	operations := request.Operations
	if len(operations) > 3 {
		return nil, rosetta.ErrInvalidRequest.RosettaError()
	}

	_, fromAddr, err := conversion.ConvertOpsToMsgs(operations)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidAddress, err.Error()).RosettaError()
	}

	memo, ok := request.Metadata["memo"]
	if !ok {
		memo = ""
	}

	defaultGas := float64(200000)

	gas := request.SuggestedFeeMultiplier
	if gas == nil {
		gas = &defaultGas
	}

	return &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			rosetta.OptionAddress: fromAddr,
			rosetta.OptionMemo:    memo,
			rosetta.OptionGas:     gas,
		},
	}, nil
}

func (on OnlineNetwork) ConstructionSubmit(ctx context.Context, request *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	res, meta, err := on.client.PostTx(txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: res,
		Metadata:              meta,
	}, nil
}
