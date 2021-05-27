package main

import (
	_ "embed"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/core/app_config"
	"github.com/cosmos/cosmos-sdk/core/cli"

	// Register Modules
	_ "github.com/cosmos/cosmos-sdk/x/authn/module"
	_ "github.com/cosmos/cosmos-sdk/x/bank/module"
)

//go:embed app_config.json
var defaultAppConfigJson []byte

func main() {
	var cfg app_config.AppConfig
	err := proto.Unmarshal(defaultAppConfigJson, &cfg)
	if err != nil {
		panic(err)
	}

	cli.Exec(cli.Config{
		DefaultAppConfig: &cfg,
		DefaultHomeDir:   "simapp",
		DefaultEnvPrefix: "SIMAPP",
	})
}
