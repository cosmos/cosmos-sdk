package init

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

func TestGenTxCmd(t *testing.T) {
	defer server.SetupViper(t)()
	defer setupClientHome(t)()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	logger := log.NewNopLogger()
	ctx := server.NewContext(cfg, logger)

	nodeID := "X"
	chainID := "X"
	ip := "X"
	valPubKey, _ := sdk.GetConsPubKeyBech32("cosmosvalconspub1zcjduepq7jsrkl9fgqk0wj3ahmfr8pgxj6vakj2wzn656s8pehh0zhv2w5as5gd80a")
	website := "http://cosmos.network"
	details := "validator details"
	identity := "that's me"
	prepareFlagsForTxCreateValidator(ctx.Config, nodeID, ip, chainID, valPubKey, website, details, identity)
	require.Equal(t, website, viper.GetString(cli.FlagWebsite))
	require.Equal(t, details, viper.GetString(cli.FlagDetails))
	require.Equal(t, identity, viper.GetString(cli.FlagIdentity))
}
