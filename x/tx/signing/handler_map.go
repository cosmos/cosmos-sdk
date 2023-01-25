package signing

import (
	"context"
	"fmt"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
)

// HandlerMap aggregates several sign mode handlers together for convenient generation of sign bytes
// based on sign mode.
type HandlerMap struct {
	signModeHandlers map[signingv1beta1.SignMode]SignModeHandler
	modes            []signingv1beta1.SignMode
}

// NewHandlerMap constructs a new sign mode handler map.
func NewHandlerMap(handlers ...SignModeHandler) *HandlerMap {
	res := &HandlerMap{
		signModeHandlers: map[signingv1beta1.SignMode]SignModeHandler{},
	}

	for _, handler := range handlers {
		mode := handler.Mode()
		res.signModeHandlers[mode] = handler
		res.modes = append(res.modes, mode)
	}

	return res
}

// SupportedModes lists the modes supported by this handler map.
func (h *HandlerMap) SupportedModes() []signingv1beta1.SignMode {
	return h.modes
}

// GetSignBytes returns the sign bytes for the transaction for the requested mode.
func (h *HandlerMap) GetSignBytes(ctx context.Context, signMode signingv1beta1.SignMode, signerData SignerData, txData TxData) ([]byte, error) {
	handler, ok := h.signModeHandlers[signMode]
	if !ok {
		return nil, fmt.Errorf("unsuppored sign mode %s", signMode)
	}

	return handler.GetSignBytes(ctx, signerData, txData)
}
