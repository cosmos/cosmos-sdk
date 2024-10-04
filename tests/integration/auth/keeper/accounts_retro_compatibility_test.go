package keeper_test

import (
	"context"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/x/accounts/accountstd"
	basev1 "cosmossdk.io/x/accounts/defaults/base/v1"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ accountstd.Interface = mockRetroCompatAccount{}

type mockRetroCompatAccount struct {
	retroCompat *authtypes.QueryLegacyAccountResponse
	address     []byte
}

func (m mockRetroCompatAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, func(ctx context.Context, req *gogotypes.Empty) (*gogotypes.Empty, error) {
		return &gogotypes.Empty{}, nil
	})
}

func (m mockRetroCompatAccount) RegisterExecuteHandlers(_ *accountstd.ExecuteBuilder) {}

func (m mockRetroCompatAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	if m.retroCompat == nil {
		return
	}
	accountstd.RegisterQueryHandler(builder, func(ctx context.Context, req *authtypes.QueryLegacyAccount) (*authtypes.QueryLegacyAccountResponse, error) {
		return m.retroCompat, nil
	})
}

func TestAuthToAccountsGRPCCompat(t *testing.T) {
	valid := &mockRetroCompatAccount{
		retroCompat: &authtypes.QueryLegacyAccountResponse{
			Account: &codectypes.Any{},
			Base: &authtypes.BaseAccount{
				Address:       "test",
				PubKey:        nil,
				AccountNumber: 10,
				Sequence:      20,
			},
		},
	}

	noInfo := &mockRetroCompatAccount{
		retroCompat: &authtypes.QueryLegacyAccountResponse{
			Account: &codectypes.Any{},
		},
	}
	noImplement := &mockRetroCompatAccount{
		retroCompat: nil,
	}

	accs := map[string]accountstd.Interface{
		"valid":        valid,
		"no_info":      noInfo,
		"no_implement": noImplement,
	}

	f := initFixture(t, accs)

	// init three accounts
	for n, a := range accs {
		_, addr, err := f.accountsKeeper.Init(f.ctx, n, []byte("me"), &gogotypes.Empty{}, nil)
		require.NoError(t, err)
		a.(*mockRetroCompatAccount).address = addr
	}

	qs := authkeeper.NewQueryServer(f.authKeeper)

	t.Run("account supports info and account query", func(t *testing.T) {
		infoResp, err := qs.AccountInfo(f.ctx, &authtypes.QueryAccountInfoRequest{
			Address: f.mustAddr(valid.address),
		})
		require.NoError(t, err)
		require.Equal(t, infoResp.Info, valid.retroCompat.Base)

		accountResp, err := qs.Account(f.ctx, &authtypes.QueryAccountRequest{
			Address: f.mustAddr(noInfo.address),
		})
		require.NoError(t, err)
		require.Equal(t, accountResp.Account, valid.retroCompat.Account)
	})

	t.Run("account only supports account query, not info", func(t *testing.T) {
		_, err := qs.AccountInfo(f.ctx, &authtypes.QueryAccountInfoRequest{
			Address: f.mustAddr(noInfo.address),
		})
		require.Error(t, err)
		require.Equal(t, status.Code(err), codes.NotFound)

		resp, err := qs.Account(f.ctx, &authtypes.QueryAccountRequest{
			Address: f.mustAddr(noInfo.address),
		})
		require.NoError(t, err)
		require.Equal(t, resp.Account, valid.retroCompat.Account)
	})

	t.Run("account does not support any retro compat", func(t *testing.T) {
		_, err := qs.AccountInfo(f.ctx, &authtypes.QueryAccountInfoRequest{
			Address: f.mustAddr(noImplement.address),
		})
		require.Error(t, err)
		require.Equal(t, status.Code(err), codes.NotFound)

		_, err = qs.Account(f.ctx, &authtypes.QueryAccountRequest{
			Address: f.mustAddr(noImplement.address),
		})

		require.Error(t, err)
		require.Equal(t, status.Code(err), codes.NotFound)
	})
}

func TestAccountsBaseAccountRetroCompat(t *testing.T) {
	f := initFixture(t, nil)
	// init a base acc
	anyPk, err := codectypes.NewAnyWithValue(secp256k1.GenPrivKey().PubKey())
	require.NoError(t, err)

	// we init two accounts to have account num not be zero.
	_, _, err = f.accountsKeeper.Init(f.ctx, "base", []byte("me"), &basev1.MsgInit{PubKey: anyPk}, nil)
	require.NoError(t, err)

	_, addr, err := f.accountsKeeper.Init(f.ctx, "base", []byte("me"), &basev1.MsgInit{PubKey: anyPk}, nil)
	require.NoError(t, err)

	// try to query it via auth
	qs := authkeeper.NewQueryServer(f.authKeeper)

	r, err := qs.Account(f.ctx, &authtypes.QueryAccountRequest{
		Address: f.mustAddr(addr),
	})
	require.NoError(t, err)
	require.NotNil(t, r.Account)

	info, err := qs.AccountInfo(f.ctx, &authtypes.QueryAccountInfoRequest{
		Address: f.mustAddr(addr),
	})
	require.NoError(t, err)
	require.NotNil(t, info.Info)
	require.Equal(t, info.Info.PubKey, anyPk)
	require.Equal(t, info.Info.AccountNumber, uint64(1))
}
