package v2

import (
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"

	"github.com/spf13/viper"
)

// AppExporter is a function that dumps all app state to
// JSON-serializable structure and returns the current validator set.
type AppExporter func(
	logger log.Logger,
	height int64,
	jailAllowedAddrs []string,
	viper *viper.Viper,
) (serverv2.ExportedApp, error)
