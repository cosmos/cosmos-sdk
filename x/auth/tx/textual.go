package tx

import (
	"context"
	"fmt"

	"cosmossdk.io/tx/textual"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"google.golang.org/protobuf/types/known/anypb"

	txsigning "cosmossdk.io/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// signModeTextualHandler defines the SIGN_MODE_TEXTUAL SignModeHandler
type signModeTextualHandler struct {
	t textual.Textual
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
	panic("SIGN_MODE_TEXTUAL needs GetSignBytesWithContext")
}

// GetSignBytesWithContext implements SignModeHandler.GetSignBytesWithContext
func (h signModeTextualHandler) GetSignBytesWithContext(ctx context.Context, mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
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

	// The first argument needs: https://github.com/cosmos/cosmos-sdk/pull/13701
	return h.t.GetSignBytes(ctx, bodyBz, authInfoBz, txsigning.SignerData{
		Address:       data.Address,
		ChainId:       data.ChainID,
		AccountNumber: data.AccountNumber,
		Sequence:      data.Sequence,
		PubKey: &anypb.Any{
			TypeUrl: pbAny.TypeUrl,
			Value:   pbAny.Value,
		},
	})
}
