package aminojson

import (
	"context"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/aminojson/internal/aminojsonpb"
)

// SignModeHandler implements the SIGN_MODE_LEGACY_AMINO_JSON signing mode.
type SignModeHandler struct {
	fileResolver signing.ProtoFileResolver
	typeResolver protoregistry.MessageTypeResolver
	encoder      Encoder
}

// SignModeHandlerOptions are the options for the SignModeHandler.
type SignModeHandlerOptions struct {
	FileResolver signing.ProtoFileResolver
	TypeResolver signing.TypeResolver
	Encoder      *Encoder
}

// NewSignModeHandler returns a new SignModeHandler.
func NewSignModeHandler(options SignModeHandlerOptions) *SignModeHandler {
	h := &SignModeHandler{}
	if options.FileResolver == nil {
		h.fileResolver = gogoproto.HybridResolver
	} else {
		h.fileResolver = options.FileResolver
	}
	if options.TypeResolver == nil {
		h.typeResolver = protoregistry.GlobalTypes
	} else {
		h.typeResolver = options.TypeResolver
	}
	if options.Encoder == nil {
		h.encoder = NewEncoder(EncoderOptions{
			FileResolver: options.FileResolver,
			TypeResolver: options.TypeResolver,
			EnumAsString: false,
		})
	} else {
		h.encoder = *options.Encoder
	}
	return h
}

// Mode implements the Mode method of the SignModeHandler interface.
func (h SignModeHandler) Mode() signing.SignMode {
	return signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

// GetSignBytes implements the GetSignBytes method of the SignModeHandler interface.
func (h SignModeHandler) GetSignBytes(_ context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	body := txData.Body

	// Get TxBody descriptor from the global registry to run unknown field checks.
	// gogoproto's init() registers all SDK proto types into protoregistry.GlobalTypes.
	txBodyMT, err := protoregistry.GlobalTypes.FindMessageByName("cosmos.tx.v1beta1.TxBody")
	if err != nil {
		return nil, fmt.Errorf("txbody descriptor not in global registry: %w", err)
	}
	_, err = signing.RejectUnknownFields(txData.BodyBytes, txBodyMT.Descriptor(), false, h.fileResolver)
	if err != nil {
		return nil, err
	}

	if len(body.ExtensionOptions) > 0 || len(body.NonCriticalExtensionOptions) > 0 {
		return nil, fmt.Errorf("%s does not support protobuf extension options: invalid request", h.Mode())
	}

	if signerData.Address == "" {
		return nil, fmt.Errorf("got empty address in %s handler: invalid request", h.Mode())
	}

	f := txData.AuthInfo.Fee

	// Convert []TxCoinData → []*basev1beta1.Coin for the amino sign fee.
	feeCoins := make([]*basev1beta1.Coin, len(f.Amount))
	for i, c := range f.Amount {
		feeCoins[i] = &basev1beta1.Coin{Denom: c.Denom, Amount: c.Amount}
	}
	fee := &aminojsonpb.AminoSignFee{
		Amount:  feeCoins,
		Gas:     f.GasLimit,
		Payer:   f.Payer,
		Granter: f.Granter,
	}

	// Convert []RawMsg → []*anypb.Any for the amino sign doc Msgs field.
	msgs := make([]*anypb.Any, len(body.Messages))
	for i, m := range body.Messages {
		msgs[i] = &anypb.Any{TypeUrl: m.TypeUrl, Value: m.Value}
	}

	signDoc := &aminojsonpb.AminoSignDoc{
		AccountNumber:    signerData.AccountNumber,
		TimeoutHeight:    body.TimeoutHeight,
		ChainId:          signerData.ChainID,
		Sequence:         signerData.Sequence,
		Memo:             body.Memo,
		Msgs:             msgs,
		Unordered:        body.Unordered,
		TimeoutTimestamp: body.TimeoutTimestamp,
		Fee:              fee,
	}

	return h.encoder.Marshal(signDoc)
}

var _ signing.SignModeHandler = (*SignModeHandler)(nil)
