package valuerenderer_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type valueRendererTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         types.Context
	queryClient banktypes.QueryClient
}

func TestValueRendererTestSuite(t *testing.T) {
	suite.Run(t, new(valueRendererTestSuite))
}

func (suite *valueRendererTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	banktypes.RegisterQueryServer(queryHelper, app.BankKeeper)
	queryClient := banktypes.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *valueRendererTestSuite) TestFormatDenomQuerierFunc() {
	metadataRegen := banktypes.Metadata{
		Name:        "Regen",
		Symbol:      "REGEN",
		Description: "The native staking token of the Regen network.",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uregen",
				Exponent: 0,
				Aliases:  []string{"microregen"},
			},
			{
				Denom:    "mregen",
				Exponent: 3,
				Aliases:  []string{"milliregen"},
			},
			{
				Denom:    "regen",
				Exponent: 6,
				Aliases:  []string{"regen"},
			},
		},
		Base:    "uregen",
		Display: "regen",
	}

	tt := []struct {
		name    string
		coin    types.Coin
		pretest func()
		expRes  string
		expErr  bool
	}{
		{
			"convert 1000000uregen to 1regen",
			types.NewCoin("uregen", types.NewInt(int64(1000000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
			},
			"1regen",
			false,
		},
		{
			"convert 23000mregen to 23regen",
			types.NewCoin("mregen", types.NewInt(int64(23000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
			},
			"23regen",
			false,
		},
		{
			"invalid coin denom",
			types.NewCoin("atom", types.NewInt(int64(23000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
			},
			"",
			true,
		},
		{
			"convert 20000000000uatom to 20000atom, multiple denoms",
			types.NewCoin("uatom", types.NewInt(int64(20000000000))),
			func() {
				metadataAtom := banktypes.Metadata{
					Name:        "Cosmos Hub Atom",
					Symbol:      "ATOM",
					Description: "The native staking token of the Cosmos Hub.",
					DenomUnits: []*banktypes.DenomUnit{
						{"uatom", uint32(0), []string{"microatom"}},
						{"matom", uint32(3), []string{"milliatom"}},
						{"atom", uint32(6), nil},
					},
					Base:    "uatom",
					Display: "atom",
				}

				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
			},
			"20,000atom",
			false,
		},
		{
			"convert 23mregen to 0,023regen",
			types.NewCoin("mregen", types.NewInt(int64(23))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
			},
			"0.023regen",
			false,
		},
	}

	for _, tc := range tt {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.pretest()

			c := types.WrapSDKContext(suite.ctx)

			dvr := valuerenderer.NewDefaultValueRenderer(func(c context.Context, baseDenom string) (banktypes.Metadata, error) {
				req := &banktypes.QueryDenomMetadataRequest{
					Denom: baseDenom,
				}

				resp, err := suite.queryClient.DenomMetadata(c, req)
				if err != nil {
					return banktypes.Metadata{}, err
				}

				return resp.Metadata, nil
			})

			res, err := dvr.Format(c, tc.coin)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Empty(res)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(res)
				suite.Require().Equal(tc.expRes, res)
			}
		})
	}
}

func TestFormatDec(t *testing.T) {
	var (
		d valuerenderer.DefaultValueRenderer
		dec types.Dec
	)

	tt := []struct {
		name   string
		pretest func()
		expRes string
	}{
		{
			"10, decimal 05",
			func() {
				dec, _ = types.NewDecFromStr("10000000.05")
			},
			"10,000,000.05",
		},
		{
			"5 mil decimal",
			func() {
				dec, _ = types.NewDecFromStr("5000000.077")
			},
			"5,000,000.077",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.pretest()
			res, err := d.Format(context.Background(), dec)
			require.NoError(t, err)
			require.Equal(t, res ,tc.expRes)
		})
	}
}

func TestFormatInt(t *testing.T) {
	var (
		d valuerenderer.DefaultValueRenderer
	)

	tt := []struct {
		name   string
		input  types.Int
		expRes string
		expErr bool
	}{
		{
			"1000000",
			types.NewInt(1000000),
			"1,000,000",
			false,
		},
		{
			"100",
			types.NewInt(100),
			"100",
			false,
		},
		{
			"23232345476756",
			types.NewInt(23232345476756),
			"23,232,345,476,756",
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.Format(context.Background(), tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}

func TestParseString(t *testing.T) {
	re := regexp.MustCompile(`(\d+)(\w+)`) // TODO test if denom starts with m or u
	dvr := valuerenderer.NewDefaultValueRenderer(func(c context.Context, denom string) (banktypes.Metadata, error) {
		return banktypes.Metadata{}, nil
	})

	tt := []struct {
		str           string
		satisfyRegExp bool
		expErr        bool
		expPanic      bool
	}{
		{"", false, true, false},
		{"10regen", true, false, false},
		{"1,000,000", false, true, true},
		{"323,000,000", false, false, true},
		{"1mregen", true, false, false},
		{"500uregen", true, false, false},
		{"1,500,000,000regen", true, false, false},
		{"394,382,328uregen", true, false, false},
	}

	for _, tc := range tt {
		t.Run(tc.str, func(t *testing.T) {
			if tc.expPanic {
				require.Panics(t, func() {
					dvr.Parse(context.Background(), tc.str)
				})
				return
			}

			x, err := dvr.Parse(context.Background(), tc.str)
			if tc.expErr {
				require.Error(t, err)
				require.Nil(t, x)
				return
			}

			require.NoError(t, err)
			if tc.satisfyRegExp {
				// should i validate coin here?
				coin, ok := x.(types.Coin)
				require.True(t, ok)
				require.NotNil(t, coin)
				require.True(t, re.MatchString(tc.str))
			} else {
				u, ok := x.(types.Uint)
				require.True(t, ok)
				require.NotNil(t, u)
			}
		})
	}
}

func TestComputeAmount(t *testing.T) {
	dvr := valuerenderer.NewDefaultValueRenderer(func(c context.Context, denom string) (banktypes.Metadata, error) {
		return banktypes.Metadata{}, nil
	})

	tt := []struct {
		name        string
		coinAmount  int64
		exponentSub int64
		expRes      string
	}{
		{"1b", 1000000000, -3, "1,000,000"},
		{"1m", 1000000, -3, "1,000"},
		{"23", 23, -3, "0.023"},
		{"1", 1, -6, "0.000001"},
		{"10", 10, 6, "10,000,000"},
		{"47", 47, 3, "47,000"},
		{"49", 49, 0, "49"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			amount := dvr.ComputeAmount(tc.coinAmount, tc.exponentSub)
			require.Equal(t, amount, tc.expRes)
		})
	}

}
