package testutil

import (
	_ "embed"

	"cosmossdk.io/core/appconfig"
)

//go:embed app.yaml
var appConfig []byte

var AppConfig = appconfig.LoadYAML(appConfig)
