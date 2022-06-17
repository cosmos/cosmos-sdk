package testutil

import (
	_ "embed"

	"cosmossdk.io/core/appconfig"
	"github.com/cosmos/cosmos-sdk/depinject"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/genutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/cosmos/cosmos-sdk/x/group/module"
	_ "github.com/cosmos/cosmos-sdk/x/mint"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

//go:embed app.yaml
var appConfig []byte

var AppConfig = depinject.Configs(
	appconfig.LoadYAML(appConfig),
	depinject.Supply(
		bank.Authority(authtypes.NewModuleAddress(govtypes.ModuleName).String()),
	),
)
