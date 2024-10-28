package authz

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/authz/types"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func setup(t *testing.T, ctx context.Context, ss store.KVStoreService, codec codec.Codec) *Account {
	t.Helper()
	deps := MakeMockDependencies(ss, codec)
	granter := "granter"

	authzAcc, err := NewAccount(deps)
	require.NoError(t, err)
	_, err = authzAcc.Init(ctx, &types.MsgInitAuthzAccount{
		Granter: granter,
	})
	require.NoError(t, err)

	return authzAcc
}

func TestGrant(t *testing.T) {
	ctx, ss := NewMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	currentTime := time.Now()
	validExpiredTime := currentTime.Add(time.Duration(time.Minute))
	validMockAuth := &types.GenericAuthoriztion{
		MsgTypeUrl: sdk.MsgTypeURL(&banktypes.MsgSend{}),
	}
	validAuthAny, err := codectypes.NewAnyWithValue(validMockAuth)
	require.NoError(t, err)
	mockGrant := types.Grant{
		Authorization: validAuthAny,
		Expiration:    &validExpiredTime,
	}

	invalidExpiredTime := currentTime.Add(time.Duration(-time.Minute))
	expiredMockGrant := types.Grant{
		Authorization: validAuthAny,
		Expiration:    &invalidExpiredTime,
	}

	testcases := []struct {
		name   string
		msg    types.MsgGrant
		expErr error
	}{
		{
			"success",
			types.MsgGrant{
				Grant:   mockGrant,
				Granter: "granter",
				Grantee: "grantee",
			},
			nil,
		},
		{
			"expired grant",
			types.MsgGrant{
				Grant:   expiredMockGrant,
				Granter: "granter",
				Grantee: "grantee",
			},
			types.ErrInvalidExpirationTime,
		},
	}

	for _, test := range testcases {
		account := setup(t, sdkCtx, ss, nil)

		_, err := account.Grant(sdkCtx, &test.msg)
		if test.expErr != nil {
			require.EqualError(t, err, test.expErr.Error(), test.name+" error not equal")
			continue
		}
		require.NoError(t, err, test.name)

		// check if state is updated
		granteeByte, err := account.addressCodec.StringToBytes(test.msg.Grantee)
		grant, err := account.Grantees.Get(sdkCtx, collections.Join(granteeByte, sdk.MsgTypeURL(&banktypes.MsgSend{})))
		require.NoError(t, err, test.name)
		require.True(t, grant.Expiration.Equal(validExpiredTime), test.name)
	}
}

func TestExecute(t *testing.T) {
	ctx, ss := NewMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	currentTime := time.Now()
	expiredTime := currentTime.Add(time.Duration(time.Minute))
	mockAuth := &types.GenericAuthoriztion{
		MsgTypeUrl: sdk.MsgTypeURL(&banktypes.MsgSend{}),
	}
	authAny, err := codectypes.NewAnyWithValue(mockAuth)
	require.NoError(t, err)
	mockGrant := types.Grant{
		Authorization: authAny,
		Expiration:    &expiredTime,
	}

	fmt.Println(mockGrant.Authorization.GetCachedValue())

	expectsentAmt := sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))
	msgSend := banktypes.MsgSend{
		FromAddress: "granter",
		ToAddress:   "random_address",
		Amount:      expectsentAmt,
	}
	msgAny, err := accountstd.PackAny(&msgSend)
	require.NoError(t, err)

	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})
	RegisterInterfaces(encCfg.InterfaceRegistry)

	account := setup(t, sdkCtx, ss, encCfg.Codec)

	// grant the account
	_, err = account.Grant(sdkCtx, &types.MsgGrant{
		Grant:   mockGrant,
		Granter: "granter",
		Grantee: "grantee",
	})
	require.NoError(t, err)

	testcases := []struct {
		name          string
		msg           types.MsgExec
		isGranteeExec bool
		expErr        error
	}{
		{
			"success -  grantee execute",
			types.MsgExec{
				Sender: "grantee",
				Msgs:   []*codectypes.Any{msgAny},
			},
			true,
			nil,
		},
		{
			"success -  granter execute",
			types.MsgExec{
				Sender: "granter",
				Msgs:   []*codectypes.Any{msgAny},
			},
			false,
			nil,
		},
	}

	for _, test := range testcases {
		_, err = account.Exec(sdkCtx, &test.msg)
		if test.expErr != nil {
			require.EqualError(t, err, test.expErr.Error(), test.name+" error not equal")
			continue
		}
		require.NoError(t, err, test.name)
	}
}
