package signing

// SignMode represents a signing mode with its own security guarantees.
// It is defined here as a standalone integer type (equivalent to the
// gogoproto-generated cosmos.tx.signing.v1beta1.SignMode) so that x/tx
// does not need to import cosmossdk.io/api/cosmos/tx/signing/v1beta1.
//
// Values are identical to the proto enum values; conversion to/from the
// gogoproto type (types/tx/signing.SignMode) is a plain integer cast.
type SignMode int32

const (
	SignMode_SIGN_MODE_UNSPECIFIED       SignMode = 0
	SignMode_SIGN_MODE_DIRECT            SignMode = 1
	SignMode_SIGN_MODE_TEXTUAL           SignMode = 2
	SignMode_SIGN_MODE_DIRECT_AUX        SignMode = 3
	SignMode_SIGN_MODE_LEGACY_AMINO_JSON SignMode = 127
	SignMode_SIGN_MODE_EIP_191           SignMode = 191
)

var signModeNames = map[SignMode]string{
	SignMode_SIGN_MODE_UNSPECIFIED:       "SIGN_MODE_UNSPECIFIED",
	SignMode_SIGN_MODE_DIRECT:            "SIGN_MODE_DIRECT",
	SignMode_SIGN_MODE_TEXTUAL:           "SIGN_MODE_TEXTUAL",
	SignMode_SIGN_MODE_DIRECT_AUX:        "SIGN_MODE_DIRECT_AUX",
	SignMode_SIGN_MODE_LEGACY_AMINO_JSON: "SIGN_MODE_LEGACY_AMINO_JSON",
	SignMode_SIGN_MODE_EIP_191:           "SIGN_MODE_EIP_191",
}

func (s SignMode) String() string {
	if name, ok := signModeNames[s]; ok {
		return name
	}
	return "SIGN_MODE_UNKNOWN"
}
