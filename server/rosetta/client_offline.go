package rosetta

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"strings"

	"github.com/tendermint/tendermint/crypto"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/coinbase/rosetta-sdk-go/types"
	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func (c *Client) OperationStatuses() []*types.OperationStatus {
	return []*types.OperationStatus{
		{
			Status:     StatusTxSuccess,
			Successful: true,
		},
		{
			Status:     StatusTxReverted,
			Successful: false,
		},
	}
}

func (c *Client) Version() string {
	return c.version
}

func (c *Client) SupportedOperations() []string {
	var supportedOperations []string
	for _, ii := range c.ir.ListImplementations("cosmos.base.v1beta1.Msg") {
		resolvedMsg, err := c.ir.Resolve(ii)
		if err != nil {
			continue
		}

		if _, ok := resolvedMsg.(sdk.Msg); ok {
			supportedOperations = append(supportedOperations, strings.TrimLeft(ii, "/"))
		}
	}

	supportedOperations = append(
		supportedOperations,
		banktypes.EventTypeCoinSpent, banktypes.EventTypeCoinReceived,
	)

	return supportedOperations
}

func (c *Client) SignedTx(_ context.Context, txBytes []byte, signatures []*types.Signature) (signedTxBytes []byte, err error) {
	return c.converter.ToSDK().SignedTx(txBytes, signatures)
}

func (c *Client) ConstructionPayload(_ context.Context, request *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error) {
	b, _ := json.Marshal(request)
	log.Printf("raw req: %s", b)
	// check if there is at least one operation
	if len(request.Operations) < 1 {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidOperation, "expected at least one operation")
	}

	tx, err := c.converter.ToSDK().UnsignedTx(request.Operations)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidOperation, err.Error())
	}

	metadata := new(ConstructionMetadata)
	if err = metadata.FromMetadata(request.Metadata); err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, err.Error())
	}
	feeAmt, err := sdk.ParseCoinsNormalized(metadata.GasPrice)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, err.Error())
	}

	//
	builder := c.txConfig.NewTxBuilder()
	_ = builder.SetMsgs(tx.GetMsgs()...)
	builder.SetFeeAmount(feeAmt)
	builder.SetGasLimit(metadata.GasLimit)
	builder.SetMemo(metadata.Memo)

	tx = builder.GetTx()

	accIdentifiers := tx.GetSigners()

	payloads := make([]*types.SigningPayload, len(accIdentifiers))
	signersData := make([]signing.SignatureV2, len(accIdentifiers))
	for i, accID := range accIdentifiers {
		// we expect pubkeys to be ordered... TODO(fdymylja): maybe make ordering not matter?
		signerData := authsigning.SignerData{
			ChainID:       metadata.ChainID,
			AccountNumber: metadata.SignersData[i].AccountNumber,
			Sequence:      metadata.SignersData[i].Sequence,
		}
		// Sign_mode_legacy_amino is being used as default here, as sign_mode_direct
		// needs the signer infos to be set before hand but rosetta doesn't have a way
		// to do this yet. To be revisited in future versions of sdk and rosetta
		signBytes, err := c.txConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signerData, tx)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "signing error: "+err.Error())
		}

		payloads[i] = &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{Address: accID.String()},
			Bytes:             crypto.Sha256(signBytes),
			SignatureType:     types.Ecdsa,
		}

		pk, err := c.converter.ToSDK().PubKey(request.PublicKeys[i])
		if err != nil {
			return nil, err
		}

		signersData[i] = signing.SignatureV2{
			PubKey:   pk,
			Data:     &signing.SingleSignatureData{},
			Sequence: metadata.SignersData[i].Sequence,
		}
	}

	// we set the signature data so we carry information regarding public key
	// then afterwards we just need to set the signed bytes
	err = builder.SetSignatures(signersData...)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	// encode tx
	encodedTx, err := c.txEncode(builder.GetTx())
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(encodedTx),
		Payloads:            payloads,
	}, nil
}

func (c *Client) PreprocessOperationsToOptions(_ context.Context, req *types.ConstructionPreprocessRequest) (response *types.ConstructionPreprocessResponse, err error) {
	if len(req.Operations) == 0 {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "no operations")
	}

	// now we need to parse the operations to cosmos sdk messages
	tx, err := c.converter.ToSDK().UnsignedTx(req.Operations)
	if err != nil {
		return nil, err
	}

	// get the signers
	signers := tx.GetSigners()
	signersStr := make([]string, len(signers))
	accountIdentifiers := make([]*types.AccountIdentifier, len(signers))

	for i, sig := range signers {
		addr := sig.String()
		signersStr[i] = addr
		accountIdentifiers[i] = &types.AccountIdentifier{
			Address: addr,
		}
	}
	// get the metadata request information
	meta := new(ConstructionPreprocessMetadata)
	err = meta.FromMetadata(req.Metadata)
	if err != nil {
		return nil, err
	}

	if meta.GasPrice == "" {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "no gas prices")
	}

	if meta.GasLimit == 0 {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "no gas limit")
	}

	// prepare the options to return
	options := &PreprocessOperationsOptionsResponse{
		ExpectedSigners: signersStr,
		Memo:            meta.Memo,
		GasLimit:        meta.GasLimit,
		GasPrice:        meta.GasPrice,
	}

	metaOptions, err := options.ToMetadata()
	if err != nil {
		return nil, err
	}
	return &types.ConstructionPreprocessResponse{
		Options:            metaOptions,
		RequiredPublicKeys: accountIdentifiers,
	}, nil
}

func (c *Client) AccountIdentifierFromPublicKey(pubKey *types.PublicKey) (*types.AccountIdentifier, error) {
	pk, err := c.converter.ToSDK().PubKey(pubKey)
	if err != nil {
		return nil, err
	}

	return &types.AccountIdentifier{
		Address: sdk.AccAddress(pk.Address()).String(),
	}, nil
}
