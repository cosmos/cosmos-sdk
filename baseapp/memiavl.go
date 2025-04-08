package baseapp

import (
	"path/filepath"

	"cosmossdk.io/log"
	"github.com/spf13/cast"

	"cosmossdk.io/store/memiavl/rootmulti"
	"github.com/crypto-org-chain/cronos/memiavl"

	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagMemIAVL             = "memiavl.enable"
	FlagAsyncCommitBuffer   = "memiavl.async-commit-buffer"
	FlagZeroCopy            = "memiavl.zero-copy"
	FlagSnapshotKeepRecent  = "memiavl.snapshot-keep-recent"
	FlagSnapshotInterval    = "memiavl.snapshot-interval"
	FlagCacheSize           = "memiavl.cache-size"
	FlagSnapshotWriterLimit = "memiavl.snapshot-writer-limit"
)

// SetupMemIAVL insert the memiavl setter in front of baseapp options, so that
// the default rootmulti store is replaced by memiavl store,
func SetupMemIAVL(
	logger log.Logger,
	appOpts servertypes.AppOptions,
	supportExportNonSnapshotVersion bool,
	baseAppOptions []func(*BaseApp),
) []func(*BaseApp) {
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	if cast.ToBool(appOpts.Get(FlagMemIAVL)) {
		opts := memiavl.Options{
			AsyncCommitBuffer:   cast.ToInt(appOpts.Get(FlagAsyncCommitBuffer)),
			ZeroCopy:            cast.ToBool(appOpts.Get(FlagZeroCopy)),
			SnapshotKeepRecent:  cast.ToUint32(appOpts.Get(FlagSnapshotKeepRecent)),
			SnapshotInterval:    cast.ToUint32(appOpts.Get(FlagSnapshotInterval)),
			SnapshotWriterLimit: cast.ToInt(appOpts.Get(FlagSnapshotWriterLimit)),
		}

		if opts.ZeroCopy {
			// it's unsafe to cache zero-copied byte slices without copying them
			sdk.SetAddrCacheEnabled(false)
		}

		// cms must be overridden before the other options, because they may use the cms,
		// make sure the cms aren't be overridden by the other options later on.
		baseAppOptions = append([]func(*BaseApp){setMemIAVL(homePath, logger, opts, supportExportNonSnapshotVersion)}, baseAppOptions...)
	}

	return baseAppOptions
}

func setMemIAVL(homePath string, logger log.Logger, opts memiavl.Options, supportExportNonSnapshotVersion bool) func(*BaseApp) {
	return func(bapp *BaseApp) {
		// trigger state-sync snapshot creation by memiavl
		opts.TriggerStateSyncExport = func(height int64) {
			go bapp.SnapshotManager().SnapshotIfApplicable(height)
		}
		cms := rootmulti.NewStore(filepath.Join(homePath, "data", "memiavl.db"), logger, supportExportNonSnapshotVersion)
		cms.SetMemIAVLOptions(opts)
		bapp.SetCMS(cms)
	}
}
