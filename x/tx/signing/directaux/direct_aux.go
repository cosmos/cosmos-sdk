package directaux

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-proto/anyutil"
	gogoproto "github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// SignModeHandler is the SIGN_MODE_DIRECT_AUX implementation of signing.SignModeHandler.
type SignModeHandler struct {
	signersContext *signing.Context
	fileResolver   signing.ProtoFileResolver
	typeResolver   protoregistry.MessageTypeResolver
}

// SignModeHandlerOptions are the options for the SignModeHandler.
type SignModeHandlerOptions struct {
	TypeResolver   protoregistry.MessageTypeResolver
	SignersContext *signing.Context
}

// NewSignModeHandler returns a new SignModeHandler.
func NewSignModeHandler(options SignModeHandlerOptions) (SignModeHandler, error) {
	h := SignModeHandler{}
	if options.SignersContext == nil {
		return h, errors.New("signers context is required")
	}
	h.signersContext = options.SignersContext
	h.fileResolver = h.signersContext.FileResolver()
	if options.TypeResolver == nil {
		h.typeResolver = protoregistry.GlobalTypes
	} else {
		h.typeResolver = options.TypeResolver
	}
	return h, nil
}

var _ signing.SignModeHandler = SignModeHandler{}

// Mode implements signing.SignModeHandler.Mode.
func (h SignModeHandler) Mode() signing.SignMode {
	return signing.SignMode_SIGN_MODE_DIRECT_AUX
}

func (h SignModeHandler) getFirstSigner(txData signing.TxData) ([]byte, error) {
	if len(txData.Body.Messages) == 0 {
		return nil, errors.New("no signer found")
	}
	// Convert RawMsg to *anypb.Any for anyutil.Unpack.
	rawMsg := txData.Body.Messages[0]
	anyMsg := &anypb.Any{TypeUrl: rawMsg.TypeUrl, Value: rawMsg.Value}
	msg, err := anyutil.Unpack(anyMsg, h.fileResolver, h.typeResolver)
	if err != nil {
		return nil, err
	}
	signer, err := h.signersContext.GetSigners(msg)
	if err != nil {
		return nil, err
	}
	return signer[0], nil
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
		var addrErr error
		feePayer, addrErr = h.signersContext.AddressCodec().BytesToString(fp)
		if addrErr != nil {
			return nil, addrErr
		}
	}
	if feePayer == signerData.Address {
		return nil, fmt.Errorf("fee payer %s cannot sign with %s: unauthorized",
			feePayer, signing.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	signDocDirectAux := &txtypes.SignDocDirectAux{
		BodyBytes:     txData.BodyBytes,
		ChainId:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
	}

	// Convert signerData.PubKey (*anypb.Any, protov2) to gogoproto Any for SignDocDirectAux.PublicKey.
	if signerData.PubKey != nil {
		signDocDirectAux.PublicKey = &gogotypes.Any{
			TypeUrl: signerData.PubKey.TypeUrl,
			Value:   signerData.PubKey.Value,
		}
	}

	return gogoproto.Marshal(signDocDirectAux)
}

// toProtov2Any converts a protov2 Any to a raw byte form for the sign doc.
// Kept as a reference; actual conversion happens inline above.
var _ = proto.Marshal // suppress unused import
