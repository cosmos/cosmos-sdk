package signing

import (
	"context"
	"fmt"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
)

type HandlerMap struct {
	signModeHandlers map[signingv1beta1.SignMode]SignModeHandler
	modes            []signingv1beta1.SignMode
}

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

func (h *HandlerMap) SupportedModes() []signingv1beta1.SignMode {
	return h.modes
}

func (h *HandlerMap) GetSignBytes(ctx context.Context, signMode signingv1beta1.SignMode, signerData SignerData, txData TxData) ([]byte, error) {
	handler, ok := h.signModeHandlers[signMode]
	if !ok {
		return nil, fmt.Errorf("unsuppored sign mode %s", signMode)
	}

	return handler.GetSignBytes(ctx, signerData, txData)
}
