package valuerenderer_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type valueRendererTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         types.Context
	queryClient banktypes.QueryClient
	printer     *message.Printer
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
	suite.printer = message.NewPrinter(language.English)
}

func (suite *valueRendererTestSuite) TestFormatCoinSingleMetadata() {
	var (
		req         *banktypes.QueryDenomMetadataRequest
		expMetadata banktypes.Metadata
	)

	// TODO more test cases here
	tt := []struct {
		name    string
		coin    types.Coin
		pretest func()
		expRes  string
		expFail bool
		expErr  bool
	}{
		{
			"convert 23000000uregen to 23regen",
			types.NewCoin("uregen", types.NewInt(int64(23000000))),
			func() {
				expMetadata = banktypes.Metadata{
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
							Aliases:  []string{"REGEN"},
						},
					},
					Base:    "uregen",
					Display: "regen",
				}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, expMetadata)
				req = &banktypes.QueryDenomMetadataRequest{
					Denom: expMetadata.Base,
				}
			},
			"23regen",
			false,
			false,
		},
		{
			// test fails cause expMetadata.Display is set to uregen"
			"convert 23regen to 23000000uregen",
			types.NewCoin("regen", types.NewInt(int64(23000000))),
			func() {
				expMetadata = banktypes.Metadata{
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
							Aliases:  []string{"REGEN"},
						},
					},
					Base:    "uregen",
					Display: "uregen",
				}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, expMetadata)
				req = &banktypes.QueryDenomMetadataRequest{
					Denom: "regen",
				}
			},
			"23,000,000uregen",
			true,
			true,
		},
		{
			// test fails error cause metadata.Display is set to uregen
			// rpc error: code = NotFound desc = client metadata for denom mregen
			"convert 23000000mregen to 23000000000uregen",
			types.NewCoin("mregen", types.NewInt(int64(23000000))),
			func() {
				expMetadata = banktypes.Metadata{
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
							Aliases:  []string{"REGEN"},
						},
					},
					Base:    "uregen",
					Display: "uregen",
				}

				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, expMetadata)
				req = &banktypes.QueryDenomMetadataRequest{
					Denom: "mregen",
				}
			},
			"23,000,000,000uregen",
			true,
			true,
		},
	}

	for _, tc := range tt {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest() //reset
			tc.pretest()

			c := types.WrapSDKContext(suite.ctx)

			resp, err := suite.queryClient.DenomMetadata(c, req)
			if tc.expFail {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotNil(resp)
			suite.Require().Equal(resp.Metadata, expMetadata)

			dvr := valuerenderer.NewDefaultValueRenderer()
			metadatas := make([]banktypes.Metadata, 1)
			metadatas[0] = expMetadata
			dvr.SetDenomToMetadataMap(metadatas)

			res, err := dvr.Format(tc.coin)

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

func (suite *valueRendererTestSuite) TestFormatCoinMultipleMetadatas() {
	var (
		expMetadatas []banktypes.Metadata
	)

	metadataAtom := banktypes.Metadata{
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uatom",
				Exponent: 0,
				Aliases:  []string{"microatom"},
			},
			{
				Denom:    "atom",
				Exponent: 6,
				Aliases:  []string{"ATOM"},
			},
		},
		Base:    "uatom",
		Display: "atom",
	}

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
				Aliases:  []string{"mregen"},
			},
			{
				Denom:    "regen",
				Exponent: 6,
				Aliases:  []string{"REGEN"},
			},
		},
		Base:    "uregen",
		Display: "regen",
	}

	req := &banktypes.QueryDenomsMetadataRequest{
		Pagination: &query.PageRequest{
			Limit:      7,
			CountTotal: true,
		},
	}

	// TODO more test cases here e.,g invalidMetadata which to add? think about
	tt := []struct {
		name    string
		coin    types.Coin
		pretest func()
		expErr  bool
	}{
		{
			"convert 1000000uregen to 1regen",
			types.NewCoin("uregen", types.NewInt(int64(1000000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
				expMetadatas = []banktypes.Metadata{metadataAtom, metadataRegen}
			},
			false,
		},
		{
			"convert 1000000000uregen to 1000regen",
			types.NewCoin("uregen", types.NewInt(int64(1000000000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
				expMetadatas = []banktypes.Metadata{metadataAtom, metadataRegen}
			},
			false,
		},
		{
			"convert 1000000mregen to 1000regen",
			types.NewCoin("mregen", types.NewInt(int64(1000000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
				expMetadatas = []banktypes.Metadata{metadataAtom, metadataRegen}
			},
			false,
		},
		{
			"invalid expMetadata error: convert 1000000000uregen to 1000regen",
			types.NewCoin("eth", types.NewInt(int64(1000000000))),
			func() {
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadataRegen)
				expMetadatas = []banktypes.Metadata{metadataAtom, metadataRegen}
			},
			true,
		},
	}

	for _, tc := range tt {
		suite.Run(tc.name, func() {
			suite.SetupTest() //reset
			tc.pretest()

			c := types.WrapSDKContext(suite.ctx)

			resp, err := suite.queryClient.DenomsMetadata(c, req)
			suite.Require().NoError(err)
			suite.Require().Equal(resp.Metadatas, expMetadatas)

			dvr := valuerenderer.NewDefaultValueRenderer()
			dvr.SetDenomToMetadataMap(expMetadatas)

			res, err := dvr.Format(tc.coin)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Empty(res)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(res)
				metadata, err := dvr.LookupMetadataByDenom((tc.coin.Denom))
				suite.Require().NoError(err)
				suite.Require().NotNil(metadata)
				// TODO should I hardcode expRes?
				expRes := suite.printer.Sprintf("%d", dvr.ComputeAmount(tc.coin, metadata)) + metadata.Display
				suite.Require().Equal(expRes, res)
			}

		})
	}
}

// TODO address edge case  "convert 23mregen to 0,023regen" or not?
func TestFormatNoDenomQuery(t *testing.T) {

	tt := []struct {
		name      string
		coin      types.Coin
		metadatas []banktypes.Metadata
		expRes    string
		expErr    bool
	}{
		{
			"convert 1000mregen to 1000000uregen",
			types.NewCoin("mregen", types.NewInt(int64(1000))),
			[]banktypes.Metadata{
				{
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
							Aliases:  []string{"REGEN"},
						},
					},
					Base:    "uregen",
					Display: "uregen",
				},
			},
			"1,000,000uregen",
			false,
		},
		{
			"convert 23000mregen to 23regen",
			types.NewCoin("mregen", types.NewInt(int64(23000))),
			[]banktypes.Metadata{
				{
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
							Aliases:  []string{"REGEN"},
						},
					},
					Base:    "uregen",
					Display: "regen",
				},
			},
			"23regen",
			false,
		},
		{
			"convert 23000000uregen to 23regen, multiple denoms",
			types.NewCoin("mregen", types.NewInt(int64(23000))),
			[]banktypes.Metadata{
				{
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
							Aliases:  []string{"REGEN"},
						},
					},
					Base:    "uregen",
					Display: "regen",
				},
				{
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
				},
			},
			"23regen",
			false,
		},
		{
			"invalid denom",
			types.NewCoin("mregen", types.NewInt(int64(23000))),
			[]banktypes.Metadata{
				{
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
				},
			},
			"",
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dvr := valuerenderer.NewDefaultValueRenderer()
			dvr.SetDenomToMetadataMap(tc.metadatas)

			res, err := dvr.Format(tc.coin)
			if tc.expErr {
				require.Error(t, err)
				require.Empty(t, res)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				require.Equal(t, tc.expRes, res)
			}
		})
	}
}

func TestFormatDec(t *testing.T) {
	var (
		d valuerenderer.DefaultValueRenderer
	)
	// TODO add more cases and error cases

	tt := []struct {
		name   string
		input  types.Dec
		expRes string
		expErr bool
	}{
		{
			"10 thousands decimal",
			types.NewDecFromIntWithPrec(types.NewInt(1000000), 2), // 10000.000000000000000000
			"10,000.000000000000000000",
			false,
		},
		{
			"10 mil decimal",
			types.NewInt(10000000).ToDec(),
			"10,000,000.000000000000000000",
			false,
		},

		//{"invalid string input panic", "qwerty", "", true, true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}

func TestFormatInt(t *testing.T) {
	var (
		d valuerenderer.DefaultValueRenderer
	)
	// TODO add more cases and error cases
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

		//{"invalid string input panic", "qwerty", "", true, true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}

// TODO add more test cases

func TestParseString(t *testing.T) {
	re := regexp.MustCompile(`\d+[mu]?regen`)
	dvr := valuerenderer.NewDefaultValueRenderer()

	tt := []struct {
		str           string
		satisfyRegExp bool
		expErr        bool
	}{
		{"", false, true},
		{"10regen", true, false},
		{"1,000,000", false, false},
		{"323,000,000", false, false},
		{"1mregen", true, false},
		{"500uregen", true, false},
		{"1,500,000,000regen", true, false},
		{"394,382,328uregen", true, false},
	}

	for _, tc := range tt {
		t.Run(tc.str, func(t *testing.T) {
			x, err := dvr.Parse(tc.str)
			if tc.expErr {
				require.Error(t, err)
				require.Nil(t, x)
				return
			}

			if tc.satisfyRegExp {
				require.NoError(t, err)
				coin, ok := x.(types.Coin)
				require.True(t, ok)
				require.NotNil(t, coin)
				require.True(t, re.MatchString(tc.str))
			} else {
				require.NoError(t, err)
				u, ok := x.(types.Uint)
				require.True(t, ok)
				require.NotNil(t, u)
			}
		})
	}
}


/*
func (suite *IntegrationTestSuite) getTestMetadata() []types.Metadata {
	return []types.Metadata{{
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*types.DenomUnit{
			{"uatom", uint32(0), []string{"microatom"}},
			{"matom", uint32(3), []string{"milliatom"}},
			{"atom", uint32(6), nil},
		},
		Base:    "uatom",
		Display: "atom",
	},
		{
			Name:        "Token",
			Symbol:      "TOKEN",
			Description: "The native staking token of the Token Hub.",
			DenomUnits: []*types.DenomUnit{
				{"1token", uint32(5), []string{"decitoken"}},
				{"2token", uint32(4), []string{"centitoken"}},
				{"3token", uint32(7), []string{"dekatoken"}},
			},
			Base:    "utoken",
			Display: "token",
		},
	}
}


*/
