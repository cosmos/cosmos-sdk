package exported

import "context"

// AppVersionModifier defines the interface fulfilled by BaseApp
// which allows getting and setting it's appVersion field. This
// in turn updates the consensus params that are sent to the
// consensus engine in EndBlock
type AppVersionModifier interface {
	SetAppVersion(context.Context, uint64) error
	AppVersion(context.Context) uint64
}
