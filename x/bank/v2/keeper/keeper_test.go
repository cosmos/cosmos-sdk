package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/bank/v2/keeper"
	banktestutil "cosmossdk.io/x/bank/v2/testutil"
	"cosmossdk.io/x/bank/v2/types"
	banktypes "cosmossdk.io/x/bank/v2/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	fooDenom = "foo"
	barDenom = "bar"
)

var (
	burnerAcc = authtypes.NewEmptyModuleAccount(authtypes.Burner, authtypes.Burner, authtypes.Staking)
	mintAcc   = authtypes.NewEmptyModuleAccount(banktypes.MintModuleName, authtypes.Minter)

	accAddrs = []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
		sdk.AccAddress([]byte("addr3_______________")),
		sdk.AccAddress([]byte("addr4_______________")),
		sdk.AccAddress([]byte("addr5_______________")),
	}
)

func newFooCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(fooDenom, amt)
}

func newBarCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(barDenom, amt)
}

type KeeperTestSuite struct {
	suite.Suite

	ctx          context.Context
	bankKeeper   keeper.Keeper
	addressCodec address.Codec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})

	env := runtime.NewEnvironment(runtime.NewKVStoreService(key), coretesting.NewNopLogger())

	ac := codectestutil.CodecOptions{}.GetAddressCodec()
	authority := authtypes.NewModuleAddress("gov")

	suite.ctx = ctx
	suite.bankKeeper = *keeper.NewKeeper(
		authority,
		ac,
		env,
		encCfg.Codec,
	)
	suite.addressCodec = ac
}

func (suite *KeeperTestSuite) TestSendCoins_Acount_To_Account() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	// Try send with empty balances
	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt)
	require.Error(err)

	// Set balances for acc0 and then try send to acc1
	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))
	require.NoError(suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt))

	// Check balances
	acc0FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], fooDenom)
	require.Equal(acc0FooBalance.Amount, math.NewInt(90))
	acc0BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], barDenom)
	require.Equal(acc0BarBalance.Amount, math.NewInt(40))
	acc1FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], fooDenom)
	require.Equal(acc1FooBalance.Amount, math.NewInt(10))
	acc1BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], barDenom)
	require.Equal(acc1BarBalance.Amount, math.NewInt(10))
}

func (suite *KeeperTestSuite) TestSendCoins_Acount_To_Module() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	// Try send with empty balances
	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], burnerAcc.GetAddress(), sendAmt)
	require.Error(err)

	// Set balances for acc0 and then try send to acc1
	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))
	require.NoError(suite.bankKeeper.SendCoins(ctx, accAddrs[0], burnerAcc.GetAddress(), sendAmt))

	// Check balances
	acc0FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], fooDenom)
	require.Equal(acc0FooBalance.Amount, math.NewInt(90))
	acc0BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], barDenom)
	require.Equal(acc0BarBalance.Amount, math.NewInt(40))
	burnerFooBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), fooDenom)
	require.Equal(burnerFooBalance.Amount, math.NewInt(10))
	burnerBarBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), barDenom)
	require.Equal(burnerBarBalance.Amount, math.NewInt(10))
}

func (suite *KeeperTestSuite) TestSendCoins_Module_To_Account() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	require.NoError(suite.bankKeeper.MintCoins(ctx, mintAcc.GetAddress(), balances))

	// Try send from burner module
	err := suite.bankKeeper.SendCoins(ctx, burnerAcc.GetAddress(), accAddrs[4], balances)
	require.Error(err)

	// Send from mint module
	err = suite.bankKeeper.SendCoins(ctx, mintAcc.GetAddress(), accAddrs[4], balances)
	require.NoError(err)

	// Check balances
	acc4FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[4], fooDenom)
	require.Equal(acc4FooBalance.Amount, math.NewInt(100))
	acc4BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[4], barDenom)
	require.Equal(acc4BarBalance.Amount, math.NewInt(50))
	mintFooBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), fooDenom)
	require.Equal(mintFooBalance.Amount, math.NewInt(0))
	mintBarBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), barDenom)
	require.Equal(mintBarBalance.Amount, math.NewInt(0))
}

func (suite *KeeperTestSuite) TestSendCoins_Module_To_Module() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	require.NoError(suite.bankKeeper.MintCoins(ctx, mintAcc.GetAddress(), balances))

	// Try send from burner module
	err := suite.bankKeeper.SendCoins(ctx, burnerAcc.GetAddress(), mintAcc.GetAddress(), sdk.NewCoins(newFooCoin(100), newBarCoin(50)))
	require.Error(err)

	// Send from mint module to burn module
	err = suite.bankKeeper.SendCoins(ctx, mintAcc.GetAddress(), burnerAcc.GetAddress(), sdk.NewCoins(newFooCoin(100), newBarCoin(50)))
	require.NoError(err)

	// Check balances
	burnerFooBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), fooDenom)
	require.Equal(burnerFooBalance.Amount, math.NewInt(100))
	burnerBarBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), barDenom)
	require.Equal(burnerBarBalance.Amount, math.NewInt(50))
	mintFooBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), fooDenom)
	require.Equal(mintFooBalance.Amount, math.NewInt(0))
	mintBarBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), barDenom)
	require.Equal(mintBarBalance.Amount, math.NewInt(0))
}

func (suite *KeeperTestSuite) TestSendCoins_WithRestriction() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))

	// Add first restriction
	addrRestrictFunc := func(ctx context.Context, from, to []byte, amount sdk.Coins) ([]byte, error) {
		if bytes.Equal(from, to) {
			return nil, fmt.Errorf("Can not send to same address")
		}
		return to, nil
	}
	suite.bankKeeper.AppendGlobalSendRestriction(addrRestrictFunc)

	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[0], sendAmt)
	require.Error(err)
	require.Contains(err.Error(), "Can not send to same address")

	// Add second restriction
	amtRestrictFunc := func(ctx context.Context, from, to []byte, amount sdk.Coins) ([]byte, error) {
		if len(amount) > 1 {
			return nil, fmt.Errorf("Allow only one denom per one send")
		}
		return to, nil
	}
	suite.bankKeeper.AppendGlobalSendRestriction(amtRestrictFunc)

	// Pass the 1st but failt at the 2nd
	err = suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt)
	require.Error(err)
	require.Contains(err.Error(), "Allow only one denom per one send")

	// Pass both 2 restrictions
	err = suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sdk.NewCoins(newFooCoin(10)))
	require.NoError(err)

	// Check balances
	acc0FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], fooDenom)
	require.Equal(acc0FooBalance.Amount, math.NewInt(90))
	acc0BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], barDenom)
	require.Equal(acc0BarBalance.Amount, math.NewInt(50))
	acc1FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], fooDenom)
	require.Equal(acc1FooBalance.Amount, math.NewInt(10))
	acc1BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], barDenom)
	require.Equal(acc1BarBalance.Amount, math.ZeroInt())
}

func (s *KeeperTestSuite) TestCreateDenom() {
	require := s.Require()

	var (
		primaryDenom            = "foo"
		secondaryDenom          = "bar"
		defaultDenomCreationFee = banktypes.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, math.NewInt(50_000_000)))}
		twoDenomCreationFee     = banktypes.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, math.NewInt(50_000_000)), sdk.NewCoin(secondaryDenom, math.NewInt(50_000_000)))}
		nilCreationFee          = banktypes.Params{DenomCreationFee: nil}
		largeCreationFee        = banktypes.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, math.NewInt(5_000_000_000)))}
	)

	for _, tc := range []struct {
		desc             string
		denomCreationFee banktypes.Params
		setup            func()
		subdenom         string
		valid            bool
	}{
		{
			desc:             "subdenom too long",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:            false,
		},
		{
			desc:             "subdenom and creator pair already exists",
			denomCreationFee: defaultDenomCreationFee,
			setup: func() {
				_, err := s.bankKeeper.CreateDenom(s.ctx, accAddrs[0].String(), "bitcoin")
				s.Require().NoError(err)
			},
			subdenom: "bitcoin",
			valid:    false,
		},
		{
			desc:             "success case: defaultDenomCreationFee",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "evmos",
			valid:            true,
		},
		{
			desc:             "success case: twoDenomCreationFee",
			denomCreationFee: twoDenomCreationFee,
			subdenom:         "catcoin",
			valid:            true,
		},
		{
			desc:             "success case: nilCreationFee",
			denomCreationFee: nilCreationFee,
			subdenom:         "czcoin",
			valid:            true,
		},
		{
			desc:             "account doesn't have enough to pay for denom creation fee",
			denomCreationFee: largeCreationFee,
			subdenom:         "tooexpensive",
			valid:            false,
		},
		{
			desc:             "subdenom having invalid characters",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "bit/***///&&&/coin",
			valid:            false,
		},
	} {
		s.SetupTest()
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			if tc.setup != nil {
				tc.setup()
			}
			require.NoError(banktestutil.FundAccount(s.ctx, s.bankKeeper, accAddrs[0], twoDenomCreationFee.DenomCreationFee))
			require.NoError(s.bankKeeper.SetParams(s.ctx, tc.denomCreationFee))
			denomCreationFee := s.bankKeeper.GetParams(s.ctx).DenomCreationFee
			s.Require().Equal(tc.denomCreationFee.DenomCreationFee, denomCreationFee)

			// note balance, create a tokenfactory denom, then note balance again
			preCreateBalance := s.bankKeeper.GetAllBalances(s.ctx, accAddrs[0])
			newDenom, err := s.bankKeeper.CreateDenom(s.ctx, accAddrs[0].String(), tc.subdenom)
			postCreateBalance := s.bankKeeper.GetAllBalances(s.ctx, accAddrs[0])
			if tc.valid {
				s.Require().NoError(err)
				s.Require().True(preCreateBalance.Sub(postCreateBalance...).Equal(denomCreationFee))

				// Make sure that the admin is set correctly
				authority, err := s.bankKeeper.GetAuthorityMetadata(s.ctx, newDenom)

				s.Require().NoError(err)
				s.Require().Equal(accAddrs[0].String(), authority.Admin)

				// Make sure that the denom metadata is initialized correctly
				metadata, found := s.bankKeeper.GetDenomMetaData(s.ctx, newDenom)
				s.Require().True(found)
				s.Require().Equal(banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{{
						Denom:    newDenom,
						Exponent: 0,
					}},
					Base:    newDenom,
					Display: newDenom,
					Name:    newDenom,
					Symbol:  newDenom,
				}, metadata)
			} else {
				s.Require().Error(err)
				// Ensure we don't charge if we expect an error
				s.Require().True(preCreateBalance.Equal(postCreateBalance))
			}
		})
	}
}

func (s *KeeperTestSuite) TestCreateDenom_GasConsume() {
	// It's hard to estimate exactly how much gas will be consumed when creating a
	// denom, because besides consuming the gas specified by the params, the keeper
	// also does a bunch of other things that consume gas.
	//
	// Rather, we test whether the gas consumed is within a range. Specifically,
	// the range [gasConsume, gasConsume + offset]. If the actual gas consumption
	// falls within the range for all test cases, we consider the test passed.
	//
	// In experience, the total amount of gas consumed should consume be ~30k more
	// than the set amount.
	const offset = 50000

	for _, tc := range []struct {
		desc       string
		gasConsume uint64
	}{
		{
			desc:       "gas consume zero",
			gasConsume: 0,
		},
		{
			desc:       "gas consume 1,000,000",
			gasConsume: 1_000_000,
		},
		{
			desc:       "gas consume 10,000,000",
			gasConsume: 10_000_000,
		},
		{
			desc:       "gas consume 25,000,000",
			gasConsume: 25_000_000,
		},
		{
			desc:       "gas consume 50,000,000",
			gasConsume: 50_000_000,
		},
		{
			desc:       "gas consume 200,000,000",
			gasConsume: 200_000_000,
		},
	} {
		s.SetupTest()
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// set params with the gas consume amount
			s.Require().NoError(s.bankKeeper.SetParams(s.ctx, banktypes.NewParams(nil, tc.gasConsume)))

			// amount of gas consumed prior to the denom creation
			gasConsumedBefore := s.bankKeeper.Environment.GasService.GasMeter(s.ctx).Consumed()

			// create a denom
			_, err := s.bankKeeper.CreateDenom(s.ctx, accAddrs[0].String(), "test")
			s.Require().NoError(err)

			// amount of gas consumed after the denom creation
			gasConsumedAfter := s.bankKeeper.Environment.GasService.GasMeter(s.ctx).Consumed()

			// the amount of gas consumed must be within the range
			gasConsumed := gasConsumedAfter - gasConsumedBefore
			s.Require().Greater(gasConsumed, tc.gasConsume)
			s.Require().Less(gasConsumed, tc.gasConsume+offset)
		})
	}
}

func (s *KeeperTestSuite) TestMintHandler() {
	s.SetupTest()
	require := s.Require()
	s.bankKeeper.SetParams(s.ctx, types.Params{
		DenomCreationFee: sdk.NewCoins(sdk.NewCoin(fooDenom, math.NewInt(10))),
	})
	handler := keeper.NewHandlers(&s.bankKeeper)
	require.NoError(banktestutil.FundAccount(s.ctx, s.bankKeeper, accAddrs[0], sdk.NewCoins(sdk.NewCoin(fooDenom, math.NewInt(100)))))

	resp, err := handler.MsgCreateDenom(s.ctx, &types.MsgCreateDenom{
		Sender:   accAddrs[0].String(),
		Subdenom: "test",
	})
	require.NoError(err)

	newDenom := resp.NewTokenDenom
	authority := authtypes.NewModuleAddress("gov")

	for _, tc := range []struct {
		desc   string
		msg    *types.MsgMint
		expErr bool
	}{
		{
			desc: "Mint bar denom, valid",
			msg: &types.MsgMint{
				Authority: authority.String(),
				ToAddress: accAddrs[1].String(),
				Amount:    sdk.NewCoin(barDenom, math.NewInt(100)),
			},
			expErr: false,
		},
		{
			desc: "Mint bar denom, invalid authority",
			msg: &types.MsgMint{
				Authority: authority.String() + "s",
				ToAddress: accAddrs[1].String(),
				Amount:    sdk.NewCoin(barDenom, math.NewInt(100)),
			},
			expErr: true,
		},
		{
			desc: "Mint tokenfatory denom, valid",
			msg: &types.MsgMint{
				Authority: accAddrs[0].String(),
				ToAddress: accAddrs[1].String(),
				Amount:    sdk.NewCoin(newDenom, math.NewInt(100)),
			},
			expErr: false,
		},
		{
			desc: "Mint tokenfatory denom, invalid admin",
			msg: &types.MsgMint{
				Authority: authority.String(),
				ToAddress: accAddrs[1].String(),
				Amount:    sdk.NewCoin(newDenom, math.NewInt(100)),
			},
			expErr: true,
		},
		{
			desc: "Mint tokenfatory denom, denom not created",
			msg: &types.MsgMint{
				Authority: accAddrs[0].String(),
				ToAddress: accAddrs[1].String(),
				Amount:    sdk.NewCoin(newDenom+"s", math.NewInt(100)),
			},
			expErr: true,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			_, err := handler.MsgMint(s.ctx, tc.msg)
			if tc.expErr {
				require.Error(err)
			} else {
				require.NoError(err)
				// Check ToAddress balance after
				toAddr, err := s.addressCodec.StringToBytes(tc.msg.ToAddress)
				require.NoError(err)
				balance := s.bankKeeper.GetBalance(s.ctx, toAddr, tc.msg.Amount.Denom)
				require.Equal(balance, tc.msg.Amount)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBurnHandler() {
	s.SetupTest()
	require := s.Require()
	s.bankKeeper.SetParams(s.ctx, types.Params{
		DenomCreationFee: sdk.NewCoins(sdk.NewCoin(fooDenom, math.NewInt(10))),
	})
	handler := keeper.NewHandlers(&s.bankKeeper)
	require.NoError(banktestutil.FundAccount(s.ctx, s.bankKeeper, accAddrs[0], sdk.NewCoins(sdk.NewCoin(fooDenom, math.NewInt(100)))))

	resp, err := handler.MsgCreateDenom(s.ctx, &types.MsgCreateDenom{
		Sender:   accAddrs[0].String(),
		Subdenom: "test",
	})
	require.NoError(err)

	newDenom := resp.NewTokenDenom
	authority := authtypes.NewModuleAddress("gov")

	_, err = handler.MsgMint(s.ctx, &types.MsgMint{
		Authority: accAddrs[0].String(),
		ToAddress: accAddrs[0].String(),
		Amount:    sdk.NewCoin(newDenom, math.NewInt(100)),
	})
	require.NoError(err)

	for _, tc := range []struct {
		desc   string
		msg    *types.MsgBurn
		expErr bool
	}{
		{
			desc: "Burn foo denom, valid",
			msg: &types.MsgBurn{
				Authority:       authority.String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(fooDenom, math.NewInt(50)),
			},
			expErr: false,
		},
		{
			desc: "Burn foo denom, invalid authority",
			msg: &types.MsgBurn{
				Authority:       accAddrs[0].String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(fooDenom, math.NewInt(50)),
			},
			expErr: true,
		},
		{
			desc: "Burn foo denom, insufficient funds",
			msg: &types.MsgBurn{
				Authority:       authority.String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(fooDenom, math.NewInt(200)),
			},
			expErr: true,
		},
		{
			desc: "Burn bar denom, invalid denom",
			msg: &types.MsgBurn{
				Authority:       authority.String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(barDenom, math.NewInt(50)),
			},
			expErr: true,
		},
		{
			desc: "Burn tokenfactory denom, valid",
			msg: &types.MsgBurn{
				Authority:       accAddrs[0].String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(newDenom, math.NewInt(50)),
			},
			expErr: false,
		},
		{
			desc: "Burn tokenfactory denom, invalid admin",
			msg: &types.MsgBurn{
				Authority:       authority.String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(newDenom, math.NewInt(50)),
			},
			expErr: true,
		},
		{
			desc: "Burn tokenfactory denom, insufficient funds",
			msg: &types.MsgBurn{
				Authority:       accAddrs[0].String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(newDenom, math.NewInt(150)),
			},
			expErr: true,
		},
		{
			desc: "Burn tokenfactory denom, token not exist",
			msg: &types.MsgBurn{
				Authority:       authority.String(),
				BurnFromAddress: accAddrs[0].String(),
				Amount:          sdk.NewCoin(newDenom+"s", math.NewInt(50)),
			},
			expErr: true,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// Get balance before burn
			fromAddr, err := s.addressCodec.StringToBytes(tc.msg.BurnFromAddress)
			require.NoError(err)

			beforeBalances := s.bankKeeper.GetAllBalances(s.ctx, fromAddr)
			_, err = handler.MsgBurn(s.ctx, tc.msg)
			if tc.expErr {
				require.Error(err)
			} else {
				require.NoError(err)
				// Check ToAddress balance after
				afterBalances := s.bankKeeper.GetAllBalances(s.ctx, fromAddr)
				require.Equal(beforeBalances.Sub(afterBalances...), sdk.NewCoins(tc.msg.Amount))
			}
		})
	}
}

func (s *KeeperTestSuite) TestSendHandler_TokenfactoryDenom() {
	s.SetupTest()
	require := s.Require()
	s.bankKeeper.SetParams(s.ctx, types.Params{
		DenomCreationFee: sdk.NewCoins(sdk.NewCoin(fooDenom, math.NewInt(10))),
	})
	handler := keeper.NewHandlers(&s.bankKeeper)
	require.NoError(banktestutil.FundAccount(s.ctx, s.bankKeeper, accAddrs[0], sdk.NewCoins(sdk.NewCoin(fooDenom, math.NewInt(100)))))

	resp, err := handler.MsgCreateDenom(s.ctx, &types.MsgCreateDenom{
		Sender:   accAddrs[0].String(),
		Subdenom: "test",
	})
	require.NoError(err)

	newDenom := resp.NewTokenDenom

	_, err = handler.MsgMint(s.ctx, &types.MsgMint{
		Authority: accAddrs[0].String(),
		ToAddress: accAddrs[0].String(),
		Amount:    sdk.NewCoin(newDenom, math.NewInt(100)),
	})
	require.NoError(err)

	for _, tc := range []struct {
		desc   string
		msg    *types.MsgSend
		expErr bool
	}{
		{
			desc: "valid",
			msg: &types.MsgSend{
				FromAddress: accAddrs[0].String(),
				ToAddress:   accAddrs[1].String(),
				Amount:      sdk.NewCoins(sdk.NewCoin(newDenom, math.NewInt(50))),
			},
			expErr: false,
		},
		{
			desc: "insufficient funds",
			msg: &types.MsgSend{
				FromAddress: accAddrs[0].String(),
				ToAddress:   accAddrs[1].String(),
				Amount:      sdk.NewCoins(sdk.NewCoin(newDenom, math.NewInt(150))),
			},
			expErr: true,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			_, err := handler.MsgSend(s.ctx, tc.msg)
			if tc.expErr {
				require.Error(err)
			} else {
				require.NoError(err)
				// Check ToAddress balances after
				toAddr, err := s.addressCodec.StringToBytes(tc.msg.ToAddress)
				require.NoError(err)
				balances := s.bankKeeper.GetAllBalances(s.ctx, toAddr)
				require.Equal(balances, tc.msg.Amount)
			}
		})
	}
}
