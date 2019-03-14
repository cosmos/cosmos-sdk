package init

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/libs/log"
	"testing"
)

func TestGenTxCmd(t *testing.T) {
	defer server.SetupViper(t)()
	defer setupClientHome(t)()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)

	ctx := server.NewContext(cfg, logger)
	cdc := app.MakeCodec()

	cmd := InitCmd(ctx, cdc)
	require.NoError(t, cmd.RunE(nil, []string{"gaianode-test"}))

	nodeID, valPubKey, err := InitializeNodeValidatorFiles(cfg)
	//cmd2 := GenTxCmd(ctx, cdc)

	viper.Set(cli.FlagIP, "1.2.3.4")
	viper.Set(cli.FlagMoniker, "gaianode-test")
	viper.Set(client.FlagName, "gaianode-test")
	viper.Set(cli.FlagNodeID, nodeID)
	viper.Set(cli.FlagPubKey, valPubKey)
	viper.Set(cli.FlagWebsite, "not.work.ing")
	//require.NoError(t, cmd2.RunE(nil, nil))
}
