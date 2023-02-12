package std

import (
	"fmt"

	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/textual"
)

// SignModeOptions are options for configuring the standard sign mode handler map.
type SignModeOptions struct {
	// CoinMetadataQueryFn is the CoinMetadataQueryFn required for SIGN_MODE_TEXTUAL.
	CoinMetadataQueryFn textual.CoinMetadataQueryFn
}

// HandlerMap returns a sign mode handler map that Cosmos SDK apps can use out
// of the box to support all "standard" sign modes.
func (s SignModeOptions) HandlerMap() (*signing.HandlerMap, error) {
	if s.CoinMetadataQueryFn == nil {
		return nil, fmt.Errorf("missing %T needed for SIGN_MODE_TEXTUAL", s.CoinMetadataQueryFn)
	}

	return signing.NewHandlerMap(
		direct.SignModeHandler{},
		textual.NewSignModeHandler(s.CoinMetadataQueryFn),
	), nil
}
