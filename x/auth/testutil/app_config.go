package testutil

import (
	_ "embed"

	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/module"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	_ "github.com/cosmos/cosmos-sdk/x/genutil"
	_ "github.com/cosmos/cosmos-sdk/x/gov"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"

	"cosmossdk.io/core/appconfig"
)

//go:embed app.yaml
var appConfig []byte

var AppConfig = appconfig.LoadYAML(appConfig)
