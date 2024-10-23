package authz

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/authz/types"
	banktypes "cosmossdk.io/x/bank/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T, ctx context.Context, ss store.KVStoreService) *Account {
	t.Helper()
	deps := makeMockDependencies(ss)
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
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	currentTime := time.Now()
	validExpiredTime := currentTime.Add(time.Duration(time.Minute))
	validMockAuth := newMockAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{}))
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
		account := setup(t, sdkCtx, ss)

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
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	currentTime := time.Now()
	expiredTime := currentTime.Add(time.Duration(time.Minute))
	mockAuth := newMockAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{}))
	authAny, err := codectypes.NewAnyWithValue(mockAuth)
	require.NoError(t, err)
	mockGrant := types.Grant{
		Authorization: authAny,
		Expiration:    &expiredTime,
	}

	expectsentAmt := sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))
	msgSend := banktypes.MsgSend{
		FromAddress: "granter",
		ToAddress:   "random_address",
		Amount:      expectsentAmt,
	}
	msgAny, err := accountstd.PackAny(&msgSend)
	require.NoError(t, err)

	account := setup(t, sdkCtx, ss)

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

		// check if state is updated
		if test.isGranteeExec {
			granteeByte, err := account.addressCodec.StringToBytes(test.msg.Sender)
			grant, err := account.Grantees.Get(sdkCtx, collections.Join(granteeByte, sdk.MsgTypeURL(&banktypes.MsgSend{})))
			require.NoError(t, err, test.name)
			authorizationAny := grant.GetAuthorization()
			mockAuth, ok := authorizationAny.GetCachedValue().(mockAuthorization)
			require.True(t, ok)
			require.True(t, mockAuth.sendAmt.Equal(expectsentAmt))
		} else {
			granteeByte, err := account.addressCodec.StringToBytes("grantee")
			grant, err := account.Grantees.Get(sdkCtx, collections.Join(granteeByte, sdk.MsgTypeURL(&banktypes.MsgSend{})))
			require.NoError(t, err, test.name)
			authorizationAny := grant.GetAuthorization()
			mockAuth, ok := authorizationAny.GetCachedValue().(mockAuthorization)
			require.True(t, ok)
			require.True(t, mockAuth.sendAmt.Equal(sdk.NewCoins()))
		}
	}
}
