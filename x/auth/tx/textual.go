package tx

import (
	"fmt"

	"cosmossdk.io/tx/textual/valuerenderer"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"google.golang.org/protobuf/types/known/anypb"

	textualv1 "cosmossdk.io/api/cosmos/msg/textual/v1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// signModeTextualHandler defines the SIGN_MODE_TEXTUAL SignModeHandler
type signModeTextualHandler struct {
	t valuerenderer.Textual
}

var _ signing.SignModeHandler = signModeTextualHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeTextualHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_TEXTUAL
}

// Modes implements SignModeHandler.Modes
func (signModeTextualHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_TEXTUAL}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (h signModeTextualHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_TEXTUAL {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_TEXTUAL, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	bodyBz := protoTx.getBodyBytes()
	authInfoBz := protoTx.getAuthInfoBytes()

	pbAny, err := codectypes.NewAnyWithValue(data.PubKey)
	if err != nil {
		return nil, err
	}

	textualData := &textualv1.TextualData{
		BodyBytes:     bodyBz,
		AuthInfoBytes: authInfoBz,
		SignerData: &signingv1beta1.SignerData{
			Address:       data.Address,
			ChainId:       data.ChainID,
			AccountNumber: data.AccountNumber,
			Sequence:      data.Sequence,
			PubKey: &anypb.Any{
				TypeUrl: pbAny.TypeUrl,
				Value:   pbAny.Value,
			},
		},
	}

	// The first argument needs: https://github.com/cosmos/cosmos-sdk/pull/13701
	return h.t.GetSignBytes(ctx, textualData)
}
