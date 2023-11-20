package exported

import sdk "github.com/cosmos/cosmos-sdk/types"

// ProtocolVersionSetter defines the interface fulfilled by BaseApp
// which allows setting it's appVersion field.
type ProtocolVersionSetter interface {
	SetAppVersion(sdk.Context, uint64)
}
