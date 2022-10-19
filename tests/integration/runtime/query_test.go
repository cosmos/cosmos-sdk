package runtime

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/module"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type fixture struct {
	ctx            sdk.Context
	appQueryClient appv1alpha1.QueryClient
}

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}

	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := simtestutil.Setup(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.TxModule(),
			configurator.ParamsModule(),
			configurator.ConsensusModule(),
			configurator.BankModule(),
			configurator.StakingModule(),
		),
		&interfaceRegistry,
	)
	assert.NilError(t, err)

	f.ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	assert.NilError(t, app.Load(true))

	queryHelper := baseapp.NewQueryServerTestHelper(f.ctx, interfaceRegistry)
	f.appQueryClient = appv1alpha1.NewQueryClient(queryHelper)

	return f
}

func TestQueryAppConfig(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	res, err := f.appQueryClient.Config(f.ctx, &appv1alpha1.QueryConfigRequest{})
	assert.NilError(t, err)
	// app config is not nil
	assert.Assert(t, res != nil && res.Config != nil)
	// has bank module
}
