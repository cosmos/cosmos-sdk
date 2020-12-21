package rosetta

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func (c *Client) OperationStatuses() []*types.OperationStatus {
	return []*types.OperationStatus{
		{
			Status:     StatusSuccess,
			Successful: true,
		},
	}
}

func (c *Client) Version() string {
	const version = "cosmos-sdk:v0.40.0-rc5/tendermint:0.34.0"
	return version
}

func (c *Client) SupportedOperations() []string {
	var supportedOperations []string
	for _, ii := range c.ir.ListImplementations("cosmos.base.v1beta1.Msg") {
		resolve, err := c.ir.Resolve(ii)
		if err == nil {
			if _, ok := resolve.(Msg); ok {
				supportedOperations = append(supportedOperations, strings.TrimLeft(ii, "/"))
			}
		}
	}

	supportedOperations = append(supportedOperations, OperationFee)

	return supportedOperations
}

func (c *Client) SignedTx(ctx context.Context, txBytes []byte, signatures []*types.Signature) (signedTxBytes []byte, err error) {
	TxConfig := c.getTxConfig()
	rawTx, err := TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}

	txBldr, _ := TxConfig.WrapTxBuilder(rawTx)

	var sigs = make([]signing.SignatureV2, len(signatures))
	for i, signature := range signatures {
		if signature.PublicKey.CurveType != "secp256k1" {
			return nil, crgerrs.ErrUnsupportedCurve
		}

		cmp, err := btcec.ParsePubKey(signature.PublicKey.Bytes, btcec.S256())
		if err != nil {
			return nil, err
		}

		compressedPublicKey := make([]byte, secp256k1.PubKeySize)
		copy(compressedPublicKey, cmp.SerializeCompressed())
		pubKey := &secp256k1.PubKey{Key: compressedPublicKey}

		accountInfo, err := c.accountInfo(ctx, sdk.AccAddress(pubKey.Address()).String(), nil)
		if err != nil {
			return nil, err
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

	if err = txBldr.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	txBytes, err = c.getTxConfig().TxEncoder()(txBldr.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

func (c *Client) ConstructionPayload(ctx context.Context, request *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error) {
	if len(request.Operations) > 3 {
		return nil, crgerrs.ErrInvalidOperation
	}

	msgs, fee, err := operationsToSdkMsgs(c.ir, request.Operations)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidOperation, err.Error())
	}

	metadata, err := getMetadataFromPayloadReq(request)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidRequest, err.Error())
	}

	txFactory := tx.Factory{}.WithAccountNumber(metadata.AccountNumber).WithChainID(metadata.ChainID).
		WithGas(metadata.Gas).WithSequence(metadata.Sequence).WithMemo(metadata.Memo).WithFees(fee.String())

	TxConfig := c.getTxConfig()
	txFactory = txFactory.WithTxConfig(TxConfig)

	txBldr, err := tx.BuildUnsignedTx(txFactory, msgs...)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	txBytes, err := TxConfig.TxEncoder()(txBldr.GetTx())
	if err != nil {
		return nil, err
	}

	accIdentifiers := getAccountIdentifiersByMsgs(msgs)

	payloads := make([]*types.SigningPayload, len(accIdentifiers))
	for i, accID := range accIdentifiers {
		payloads[i] = &types.SigningPayload{
			AccountIdentifier: accID,
			Bytes:             crypto.Sha256(signBytes),
			SignatureType:     "ecdsa",
		}
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads:            payloads,
	}, nil
}

func getAccountIdentifiersByMsgs(msgs []sdk.Msg) []*types.AccountIdentifier {
	var accIdentifiers []*types.AccountIdentifier
	for _, msg := range msgs {
		for _, signer := range msg.GetSigners() {
			accIdentifiers = append(accIdentifiers, &types.AccountIdentifier{Address: signer.String()})
		}
	}

	return accIdentifiers
}

func (c *Client) PreprocessOperationsToOptions(ctx context.Context, req *types.ConstructionPreprocessRequest) (options map[string]interface{}, err error) {
	operations := req.Operations
	if len(operations) > 3 {
		return nil, crgerrs.ErrInvalidRequest
	}

	msgs, err := ConvertOpsToMsgs(c.ir, operations)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidAddress, err.Error())
	}

	if len(msgs) < 1 || len(msgs[0].GetSigners()) < 1 {
		return nil, crgerrs.WrapError(crgerrs.ErrInterpreting, "invalid msgs from operations")
	}

	memo, ok := req.Metadata["memo"]
	if !ok {
		memo = ""
	}

	defaultGas := float64(200000)

	gas := req.SuggestedFeeMultiplier
	if gas == nil {
		gas = &defaultGas
	}

	return map[string]interface{}{
		OptionAddress: msgs[0].GetSigners()[0],
		OptionMemo:    memo,
		OptionGas:     gas,
	}, nil
}
