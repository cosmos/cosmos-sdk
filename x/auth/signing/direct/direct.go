package direct

import (
	"fmt"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// ProtoTx defines an interface which protobuf transactions must implement for
// signature verification via SignModeDirect
type ProtoTx interface {
	// GetBodyBytes returns the raw serialized bytes for TxBody
	GetBodyBytes() []byte

	// GetBodyBytes returns the raw serialized bytes for AuthInfo
	GetAuthInfoBytes() []byte
}

// ModeHandler defines the SIGN_MODE_DIRECT SignModeHandler
type ModeHandler struct{}

var _ signing.SignModeHandler = ModeHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (ModeHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_DIRECT
}

// Modes implements SignModeHandler.Modes
func (ModeHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (ModeHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_DIRECT {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_DIRECT, mode)
	}

	protoTx, ok := tx.(ProtoTx)
	if !ok {
		return nil, fmt.Errorf("can only get direct sign bytes for a ProtoTx, got %T", tx)
	}

	bodyBz := protoTx.GetBodyBytes()
	authInfoBz := protoTx.GetAuthInfoBytes()

	return SignBytes(bodyBz, authInfoBz, data.ChainID, data.AccountNumber, data.AccountSequence)
}

// SignBytes returns the SIGN_MODE_DIRECT sign bytes for the provided TxBody bytes, AuthInfo bytes, chain ID,
// account number and sequence.
func SignBytes(bodyBytes, authInfoBytes []byte, chainID string, accnum, sequence uint64) ([]byte, error) {
	signDoc := types.SignDoc{
		BodyBytes:       bodyBytes,
		AuthInfoBytes:   authInfoBytes,
		ChainId:         chainID,
		AccountNumber:   accnum,
		AccountSequence: sequence,
	}
	return signDoc.Marshal()
}
