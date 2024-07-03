package postgres

import (
	"cosmossdk.io/schema/logutil"
)

type Options struct {
	RetainDeletions bool
	Logger          logutil.Logger
}
