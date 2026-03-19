package signing

import upstream "github.com/cosmos/cosmos-sdk/x/tx/signing"

type (
	TypeResolver     = upstream.TypeResolver
	Context          = upstream.Context
	Options          = upstream.Options
	ProtoFileResolver = upstream.ProtoFileResolver
	GetSignersFunc   = upstream.GetSignersFunc
	CustomGetSigner  = upstream.CustomGetSigner
	SignModeHandler  = upstream.SignModeHandler
	SignerData       = upstream.SignerData
	TxData           = upstream.TxData
)

var NewContext = upstream.NewContext
