package signing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type HandlerMap struct {
	defaultMode      types.SignMode
	modes            []types.SignMode
	signModeHandlers map[types.SignMode]types.SignModeHandler
}

var _ types.SignModeHandler = HandlerMap{}

func NewHandlerMap(defaultMode types.SignMode, handlers []types.SignModeHandler) *HandlerMap {
	handlerMap := make(map[types.SignMode]types.SignModeHandler)
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

	return &HandlerMap{
		defaultMode:      defaultMode,
		modes:            modes,
		signModeHandlers: handlerMap,
	}
}

func (h HandlerMap) DefaultMode() types.SignMode {
	return h.defaultMode
}

func (h HandlerMap) Modes() []types.SignMode {
	return h.modes
}

func (h HandlerMap) GetSignBytes(mode types.SignMode, data types.SigningData, tx sdk.Tx) ([]byte, error) {
	handler, found := h.signModeHandlers[mode]
	if !found {
		return nil, fmt.Errorf("can't verify sign mode %s", data.Mode.String())
	}
	return handler.GetSignBytes(mode, data, tx)
}
