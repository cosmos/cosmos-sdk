package exported

import sdk "github.com/cosmos/cosmos-sdk/types"

// ProtocolVersionManager defines the interface which allows managing the appVersion field.
type ProtocolVersionManager interface {
	GetAppVersion(ctx sdk.Context) (uint64, error)
	SetAppVersion(ctx sdk.Context, version uint64) error
}
