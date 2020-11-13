package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

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
var _ crg.ConstructionAPI = SingleNetwork{}

func (sn SingleNetwork) ConstructionCombine(ctx context.Context, request *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.UnsignedTransaction)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	TxConfig := sn.client.GetTxConfig(ctx)
	rawTx, err := TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	txBldr, _ := TxConfig.WrapTxBuilder(rawTx)

	var sigs = make([]signing.SignatureV2, len(request.Signatures))
	for i, signature := range request.Signatures {
		if signature.PublicKey.CurveType != "secp256k1" {
			return nil, rosetta.ErrUnsupportedCurve.RosettaError()
		}

		cmp, err := btcec.ParsePubKey(signature.PublicKey.Bytes, btcec.S256())
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}

		compressedPublicKey := make([]byte, secp256k1.PubKeySize)
		copy(compressedPublicKey, cmp.SerializeCompressed())
		pubKey := &secp256k1.PubKey{Key: compressedPublicKey}

		accountInfo, err := sn.client.AccountInfo(ctx, sdk.AccAddress(pubKey.Address()).String(), nil)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}

		sig := signing.SignatureV2{
			PubKey: pubKey,
			Data: &signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				Signature: signature.Bytes,
			},
			Sequence: accountInfo.GetSequence(),
		}
		sigs[i] = sig
	}

	err = txBldr.SetSignatures(sigs...)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	txBytes, err = TxConfig.TxEncoder()(txBldr.GetTx())
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(txBytes),
	}, nil
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

	pk := secp256k1.PubKey{Key: compressedPublicKey}
	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: sdk.AccAddress(pk.Address()).String(),
		},
	}, nil
}

func (sn SingleNetwork) ConstructionHash(ctx context.Context, request *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
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
			rosetta.ChainID:       status.NodeInfo.Network,
			rosetta.OptionGas:     gas,
			rosetta.OptionMemo:    memo,
		},
	}

	return res, nil
}

func (sn SingleNetwork) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.Transaction)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	TxConfig := sn.client.GetTxConfig(ctx)
	rawTx, err := TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	txBldr, _ := TxConfig.WrapTxBuilder(rawTx)

	var accountIdentifierSigners []*types.AccountIdentifier
	if request.Signed {
		addrs := txBldr.GetTx().GetSigners()
		for _, addr := range addrs {
			signer := &types.AccountIdentifier{
				Address: addr.String(),
			}
			accountIdentifierSigners = append(accountIdentifierSigners, signer)
		}
	}

	return &types.ConstructionParseResponse{
		Operations:               conversion.ToOperations(rawTx.GetMsgs(), false, true),
		AccountIdentifierSigners: accountIdentifierSigners,
	}, nil
}

func (sn SingleNetwork) ConstructionPayloads(ctx context.Context, request *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	if len(request.Operations) != 2 {
		return nil, rosetta.ErrInvalidOperation.RosettaError()
	}

	if request.Operations[0].Type != rosetta.OperationSend || request.Operations[1].Type != rosetta.OperationSend {
		return nil, rosetta.WrapError(rosetta.ErrInvalidOperation, "the operations are not Transfer").RosettaError()
	}

	sendMsg, err := conversion.GetTransferTxDataFromOperations(request.Operations)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidOperation, err.Error()).RosettaError()
	}

	metadata, err := GetMetadataFromPayloadReq(request)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidRequest, err.Error()).RosettaError()
	}

	txFactory := tx.Factory{}.WithAccountNumber(metadata.AccountNumber).WithChainID(metadata.ChainID).
		WithGas(metadata.Gas).WithSequence(metadata.Sequence).WithMemo(metadata.Memo)

	TxConfig := sn.client.GetTxConfig(ctx)
	txFactory = txFactory.WithTxConfig(TxConfig)
	txBldr, err := tx.BuildUnsignedTx(txFactory, sendMsg)
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
					Address: sendMsg.FromAddress,
				},
				Bytes:         crypto.Sha256(signBytes),
				SignatureType: "ecdsa",
			},
		},
	}, nil
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
		memo = ""
	}

	defaultGas := float64(200000)
	gas := request.SuggestedFeeMultiplier
	if gas == nil {
		gas = &defaultGas
	}
	var res = &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			rosetta.OptionAddress: txData.FromAddress,
			rosetta.OptionMemo:    memo,
			rosetta.OptionGas:     gas,
		},
	}
	return res, nil
}

func (sn SingleNetwork) ConstructionSubmit(ctx context.Context, request *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	res, err := sn.client.PostTx(ctx, txBytes)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: res.TxHash,
		},
		Metadata: map[string]interface{}{
			"log": res.RawLog,
		},
	}, nil
}
