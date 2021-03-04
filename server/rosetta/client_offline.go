package rosetta

import (
	"context"
	"encoding/hex"
	"log"
	"strings"

	"github.com/tendermint/tendermint/crypto"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

func (c *Client) SignedTx(ctx context.Context, txBytes []byte, signatures []*types.Signature) (signedTxBytes []byte, err error) {
	TxConfig := c.getTxConfig()
	rawTx, err := TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}

	txBldr, err := TxConfig.WrapTxBuilder(rawTx)
	if err != nil {
		return nil, err
	}

	var sigs = make([]signing.SignatureV2, len(signatures))
	for i, signature := range signatures {
		if signature.PublicKey.CurveType != types.Secp256k1 {
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

func (c *Client) ConstructionPayload(_ context.Context, request *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error) {
	// check if there is at least one operation
	if len(request.Operations) < 1 {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidOperation, "expected at least one operation")
	}

	tx, err := c.converter.FromRosetta().OpsToUnsignedTx(request.Operations)
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
	builder := c.getTxConfig().NewTxBuilder()
	_ = builder.SetMsgs(tx.GetMsgs()...)
	builder.SetFeeAmount(feeAmt)
	builder.SetGasLimit(metadata.GasLimit)
	builder.SetMemo(metadata.Memo)

	tx = builder.GetTx()
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return nil, err
	}

	accIdentifiers := tx.GetSigners()
	rawJSONTx, err := c.clientCtx.TxConfig.TxJSONEncoder()(builder.GetTx())
	log.Printf("raw tx: %s", rawJSONTx)

	payloads := make([]*types.SigningPayload, len(accIdentifiers))
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
		signBytes, err := c.clientCtx.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signerData, tx)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "signing error: "+err.Error())
		}

		payloads[i] = &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{Address: accID.String()},
			Bytes:             crypto.Sha256(signBytes),
			SignatureType:     types.Ecdsa,
		}
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads:            payloads,
	}, nil
}

func (c *Client) PreprocessOperationsToOptions(_ context.Context, req *types.ConstructionPreprocessRequest) (optionsMeta map[string]interface{}, err error) {
	if len(req.Operations) == 0 {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "no operations")
	}

	// now we need to parse the operations to cosmos sdk messages
	tx, err := c.converter.FromRosetta().OpsToUnsignedTx(req.Operations)
	if err != nil {
		return nil, err
	}

	// get the signers
	signers := tx.GetSigners()
	signersStr := make([]string, len(signers))

	for i, sig := range signers {
		signersStr[i] = sig.String()
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

	return options.ToMetadata()
}

func (c *Client) AccountIdentifierFromPublicKey(pubKey *types.PublicKey) (*types.AccountIdentifier, error) {
	if pubKey.CurveType != "secp256k1" {
		return nil, crgerrs.WrapError(crgerrs.ErrUnsupportedCurve, "only secp256k1 supported")
	}

	cmp, err := btcec.ParsePubKey(pubKey.Bytes, btcec.S256())
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, err.Error())
	}

	compressedPublicKey := make([]byte, secp256k1.PubKeySize)
	copy(compressedPublicKey, cmp.SerializeCompressed())

	pk := secp256k1.PubKey{Key: compressedPublicKey}

	return &types.AccountIdentifier{
		Address: sdk.AccAddress(pk.Address()).String(),
	}, nil
}
