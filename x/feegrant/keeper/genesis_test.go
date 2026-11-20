package keeper_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
	feegranttestutil "github.com/cosmos/cosmos-sdk/x/feegrant/testutil"
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

// Allowances that expired before the genesis block time are dropped on import,
// instead of failing InitGenesis (and with it the chain start).
func TestInitGenesisSkipsExpiredAllowances(t *testing.T) {
	source := initFixture(t)
	source.accountKeeper.EXPECT().GetAccount(gomock.Any(), granteeAddr).Return(authtypes.NewBaseAccountWithAddress(granteeAddr)).AnyTimes()

	otherGranterAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coins := sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000)))
	now := source.ctx.BlockTime()
	soon := now.Add(time.Hour)
	oneYear := now.AddDate(1, 0, 0)

	assert.NilError(t, source.feegrantKeeper.GrantAllowance(source.ctx, granterAddr, granteeAddr, &feegrant.BasicAllowance{SpendLimit: coins, Expiration: &soon}))
	assert.NilError(t, source.feegrantKeeper.GrantAllowance(source.ctx, otherGranterAddr, granteeAddr, &feegrant.BasicAllowance{SpendLimit: coins, Expiration: &oneYear}))

	genesis, err := source.feegrantKeeper.ExportGenesis(source.ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(genesis.Allowances), 2)

	// restart after the first allowance expired but before the second one did
	target := initFixture(t)
	target.accountKeeper.EXPECT().GetAccount(gomock.Any(), granteeAddr).Return(authtypes.NewBaseAccountWithAddress(granteeAddr)).AnyTimes()
	ctx := target.ctx.WithBlockTime(soon.Add(time.Hour))

	assert.NilError(t, target.feegrantKeeper.InitGenesis(ctx, genesis))

	imported, err := target.feegrantKeeper.ExportGenesis(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(imported.Allowances), 1)
	assert.Equal(t, imported.Allowances[0].Granter, otherGranterAddr.String())
}
