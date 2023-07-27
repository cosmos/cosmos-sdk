package keeper_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"
	"cosmossdk.io/x/feegrant/module"
	feegranttestutil "cosmossdk.io/x/feegrant/testutil"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	granteePub  = secp256k1.GenPrivKey().PubKey()
	granterPub  = secp256k1.GenPrivKey().PubKey()
	granteeAddr = sdk.AccAddress(granteePub.Address())
	granterAddr = sdk.AccAddress(granterPub.Address())
)

type genesisFixture struct {
	ctx            sdk.Context
	feegrantKeeper keeper.Keeper
	accountKeeper  *feegranttestutil.MockAccountKeeper
}

func initFixture(t *testing.T) *genesisFixture {
	t.Helper()
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	ctrl := gomock.NewController(t)
	accountKeeper := feegranttestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	return &genesisFixture{
		ctx:            testCtx.Ctx,
		feegrantKeeper: keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), accountKeeper),
		accountKeeper:  accountKeeper,
	}
}

func TestImportExportGenesis(t *testing.T) {
	f := initFixture(t)

	f.accountKeeper.EXPECT().GetAccount(gomock.Any(), granteeAddr).Return(authtypes.NewBaseAccountWithAddress(granteeAddr)).AnyTimes()
	f.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	coins := sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000)))
	now := f.ctx.BlockHeader().Time
	oneYear := now.AddDate(1, 0, 0)
	msgSrvr := keeper.NewMsgServerImpl(f.feegrantKeeper)

	allowance := &feegrant.BasicAllowance{SpendLimit: coins, Expiration: &oneYear}
	err := f.feegrantKeeper.GrantAllowance(f.ctx, granterAddr, granteeAddr, allowance)
	assert.NilError(t, err)

	genesis, err := f.feegrantKeeper.ExportGenesis(f.ctx)
	assert.NilError(t, err)

	// revoke fee allowance
	_, err = msgSrvr.RevokeAllowance(f.ctx, &feegrant.MsgRevokeAllowance{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	assert.NilError(t, err)

	err = f.feegrantKeeper.InitGenesis(f.ctx, genesis)
	assert.NilError(t, err)

	newGenesis, err := f.feegrantKeeper.ExportGenesis(f.ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, genesis, newGenesis)
}

func TestInitGenesis(t *testing.T) {
	any, err := codectypes.NewAnyWithValue(&testdata.Dog{})
	assert.NilError(t, err)

	testCases := []struct {
		name          string
		feeAllowances []feegrant.Grant
		invalidAddr   bool
	}{
		{
			"invalid granter",
			[]feegrant.Grant{
				{
					Granter: "invalid granter",
					Grantee: granteeAddr.String(),
				},
			},
			true,
		},
		{
			"invalid grantee",
			[]feegrant.Grant{
				{
					Granter: granterAddr.String(),
					Grantee: "invalid grantee",
				},
			},
			true,
		},
		{
			"invalid allowance",
			[]feegrant.Grant{
				{
					Granter:   granterAddr.String(),
					Grantee:   granteeAddr.String(),
					Allowance: any,
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := initFixture(t)
			if !tc.invalidAddr {
				err := f.feegrantKeeper.InitGenesis(f.ctx, &feegrant.GenesisState{Allowances: tc.feeAllowances})
				assert.ErrorContains(t, err, "failed to get allowance: no allowance")
			} else {
				expectedErr := errors.New("decoding bech32 failed")
				err := f.feegrantKeeper.InitGenesis(f.ctx, &feegrant.GenesisState{Allowances: tc.feeAllowances})
				assert.ErrorContains(t, err, expectedErr.Error())
			}
		})
	}
}
