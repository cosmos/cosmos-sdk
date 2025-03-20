package exported

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
)

// AppVersionModifier defines the interface fulfilled by BaseApp
// which allows getting and setting its appVersion field. This
// in turn updates the consensus params that are sent to the
// consensus engine in EndBlock.
type AppVersionModifier baseapp.VersionModifier
