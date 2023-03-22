package direct_aux

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing"
)

// SignModeHandler is the SIGN_MODE_DIRECT_AUX implementation of signing.SignModeHandler.
type SignModeHandler struct {
	signersContext *signing.GetSignersContext
	fileResolver   protodesc.Resolver
	typeResolver   protoregistry.MessageTypeResolver
}

// SignModeHandlerOptions are the options for the SignModeHandler.
type SignModeHandlerOptions struct {
	// FileResolver is the protodesc.Resolver to use for resolving proto files when unpacking any messages.
	FileResolver protodesc.Resolver

	// TypeResolver is the protoregistry.MessageTypeResolver to use for resolving proto types when unpacking any messages.
	TypeResolver protoregistry.MessageTypeResolver

	// SignersContext is the signing.GetSignersContext to use for getting signers.
	SignersContext *signing.GetSignersContext
}

// NewSignModeHandler returns a new SignModeHandler.
func NewSignModeHandler(options SignModeHandlerOptions) (SignModeHandler, error) {
	h := SignModeHandler{}

	if options.FileResolver == nil {
		h.fileResolver = protoregistry.GlobalFiles
	} else {
		h.fileResolver = options.FileResolver
	}

	if options.TypeResolver == nil {
		h.typeResolver = protoregistry.GlobalTypes
	} else {
		h.typeResolver = options.TypeResolver
	}

	if options.SignersContext == nil {
		var err error
		h.signersContext, err = signing.NewGetSignersContext(signing.GetSignersOptions{ProtoFiles: h.fileResolver})
		if err != nil {
			return h, err
		}
	} else {
		h.signersContext = options.SignersContext
	}

	return h, nil
}

var _ signing.SignModeHandler = SignModeHandler{}

// Mode implements signing.SignModeHandler.Mode.
func (h SignModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX
}

// getFirstSigner returns the first signer from the first message in the tx. It replicates behavior in
// https://github.com/cosmos/cosmos-sdk/blob/4a6a1e3cb8de459891cb0495052589673d14ef51/x/auth/tx/builder.go#L142
func (h SignModeHandler) getFirstSigner(txData signing.TxData) (string, error) {
	for _, anyMsg := range txData.Body.Messages {
		msg, err := anyutil.Unpack(anyMsg, h.fileResolver, h.typeResolver)
		if err != nil {
			return "", err
		}
		signer, err := h.signersContext.GetSigners(msg)
		if err != nil {
			return "", err
		}
		return signer[0], nil
	}
	return "", fmt.Errorf("no signer found")
}

// GetSignBytes implements signing.SignModeHandler.GetSignBytes.
func (h SignModeHandler) GetSignBytes(
	_ context.Context, signerData signing.SignerData, txData signing.TxData,
) ([]byte, error) {
	feePayer := txData.AuthInfo.Fee.Payer
	if feePayer == "" {
		fp, err := h.getFirstSigner(txData)
		if err != nil {
			return nil, err
		}
		feePayer = fp
	}
	if feePayer == signerData.Address {
		return nil, fmt.Errorf("fee payer %s cannot sign with %s: unauthorized",
			feePayer, signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	signDocDirectAux := &txv1beta1.SignDocDirectAux{
		BodyBytes:     txData.BodyBytes,
		PublicKey:     signerData.PubKey,
		ChainId:       signerData.ChainId,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		Tip:           txData.AuthInfo.Tip,
	}
	return proto.Marshal(signDocDirectAux)
}
