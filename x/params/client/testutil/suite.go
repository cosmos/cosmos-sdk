package testutil

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tm-db"

	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/params/client/cli"
	"github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/testutil"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// mySubspace is a x/params subspace created for the purpose of this
// test suite.
const (
	mySubspace = "foo"
)

// myParams defines some params in the `mySubspace` subspace.
type myParams struct{}

func (p *myParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte("bar"), 1234, func(value interface{}) error { return nil }),
	}
}

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	// Create a new AppConstructor for this test suite, where we manually
	// add a subspace and `myParams` to the x/params store.
	s.cfg.AppConstructor = func(val moduletestutil.Validator) servertypes.Application {
		var (
			appBuilder   *runtime.AppBuilder
			paramsKeeper keeper.Keeper
		)
		if err := depinject.Inject(testutil.AppConfig, &appBuilder, &paramsKeeper); err != nil {
			panic(err)
		}

		// Add this test's `myParams` to the x/params store.
		paramSet := myParams{}
		subspace := paramsKeeper.Subspace(mySubspace).WithKeyTable(paramtypes.NewKeyTable().RegisterParamSet(&paramSet))

		app := appBuilder.Build(
			val.GetCtx().Logger,
			dbm.NewMemDB(),
			nil,
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
		)

		s.Require().NoError(app.Load(false))

		// Make sure not to forget to persist `myParams` into the actual store,
		// this is done in InitChain.
		app.SetInitChainer(func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
			subspace.SetParamSet(ctx, &paramSet)

			return app.InitChainer(ctx, req)
		})

		s.Require().NoError(app.LoadLatestVersion())

		return app
	}

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewQuerySubspaceParamsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{
				"foo", "bar",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"subspace":"foo","key":"bar","value":"\"1234\""}`,
		},
		{
			"text output",
			[]string{
				"foo", "bar",
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
			},
			`key: bar
subspace: foo
value: '"1234"'`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewQuerySubspaceParamsCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}
