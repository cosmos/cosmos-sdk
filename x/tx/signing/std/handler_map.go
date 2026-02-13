package std

import (
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/aminojson"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/direct"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/directaux"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/textual"
)

// SignModeOptions are options for configuring the standard sign mode handler map.
type SignModeOptions struct {
	// Textual are options for SIGN_MODE_TEXTUAL
	Textual textual.SignModeOptions
	// DirectAux are options for SIGN_MODE_DIRECT_AUX
	DirectAux directaux.SignModeHandlerOptions
	// AminoJSON are options for SIGN_MODE_LEGACY_AMINO_JSON
	AminoJSON aminojson.SignModeHandlerOptions
}

// HandlerMap returns a sign mode handler map that Cosmos SDK apps can use out
// of the box to support all "standard" sign modes.
func (s SignModeOptions) HandlerMap() (*signing.HandlerMap, error) {
	txt, err := textual.NewSignModeHandler(s.Textual)
	if err != nil {
		return nil, err
	}

	directAux, err := directaux.NewSignModeHandler(s.DirectAux)
	if err != nil {
		return nil, err
	}

	aminoJSON := aminojson.NewSignModeHandler(s.AminoJSON)

	return signing.NewHandlerMap(
		direct.SignModeHandler{},
		txt,
		directAux,
		aminoJSON,
	), nil
}
