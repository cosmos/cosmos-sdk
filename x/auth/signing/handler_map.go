package signing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

// SignModeHandlerMap is SignModeHandler that aggregates multiple SignModeHandler's into
// a single handler
type SignModeHandlerMap struct {
	defaultMode      types.SignMode
	modes            []types.SignMode
	signModeHandlers map[types.SignMode]SignModeHandler
}

var _ SignModeHandler = SignModeHandlerMap{}

// NewSignModeHandlerMap returns a new SignModeHandlerMap with the provided defaultMode and handlers
func NewSignModeHandlerMap(defaultMode types.SignMode, handlers []SignModeHandler) SignModeHandlerMap {
	handlerMap := make(map[types.SignMode]SignModeHandler)
	var modes []types.SignMode

	for _, h := range handlers {
		for _, m := range h.Modes() {
			if _, have := handlerMap[m]; have {
				panic(fmt.Errorf("duplicate sign mode handler for mode %s", m))
			}
			handlerMap[m] = h
			modes = append(modes, m)
		}
	}

	return SignModeHandlerMap{
		defaultMode:      defaultMode,
		modes:            modes,
		signModeHandlers: handlerMap,
	}
}

// DefaultMode implements SignModeHandler.DefaultMode
func (h SignModeHandlerMap) DefaultMode() types.SignMode {
	return h.defaultMode
}

// Modes implements SignModeHandler.Modes
func (h SignModeHandlerMap) Modes() []types.SignMode {
	return h.modes
}

// DefaultMode implements SignModeHandler.GetSignBytes
func (h SignModeHandlerMap) GetSignBytes(mode types.SignMode, data SignerData, tx sdk.Tx) ([]byte, error) {
	handler, found := h.signModeHandlers[mode]
	if !found {
		return nil, fmt.Errorf("can't verify sign mode %s", mode.String())
	}
	return handler.GetSignBytes(mode, data, tx)
}
