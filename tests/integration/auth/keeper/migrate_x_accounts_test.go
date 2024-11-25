package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	basev1 "cosmossdk.io/x/accounts/defaults/base/v1"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestMigrateToAccounts(t *testing.T) {
	f := initFixture(t, nil)

	// create a module account
	modAcc := &authtypes.ModuleAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       f.mustAddr([]byte("cookies")),
			PubKey:        nil,
			AccountNumber: 0,
			Sequence:      0,
		},
		Name:        "cookies",
		Permissions: nil,
	}
	updatedMod := f.authKeeper.NewAccount(f.app.Context(), modAcc)
	f.authKeeper.SetAccount(f.app.Context(), updatedMod)

	// create account
	msgSrv := authkeeper.NewMsgServerImpl(f.authKeeper)
	privKey := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(privKey.PubKey().Address())

	acc := f.authKeeper.NewAccountWithAddress(f.app.Context(), addr)
	require.NoError(t, acc.SetPubKey(privKey.PubKey()))
	f.authKeeper.SetAccount(f.app.Context(), acc)

	t.Run("account does not exist", func(t *testing.T) {
		resp, err := msgSrv.MigrateAccount(f.app.Context(), &authtypes.MsgMigrateAccount{
			Signer:         f.mustAddr([]byte("notexist")),
			AccountType:    "base",
			AccountInitMsg: nil,
		})
		require.Nil(t, resp)
		require.ErrorIs(t, err, sdkerrors.ErrUnknownAddress)
	})

	t.Run("invalid account type", func(t *testing.T) {
		resp, err := msgSrv.MigrateAccount(f.app.Context(), &authtypes.MsgMigrateAccount{
			Signer:         f.mustAddr(updatedMod.GetAddress()),
			AccountType:    "base",
			AccountInitMsg: nil,
		})
		require.Nil(t, resp)
		require.ErrorContains(t, err, "only BaseAccount can be migrated")
	})

	t.Run("success", func(t *testing.T) {
		pk, err := codectypes.NewAnyWithValue(privKey.PubKey())
		require.NoError(t, err)

		migrateMsg := &basev1.MsgInit{
			PubKey:       pk,
			InitSequence: 100,
		}

		initMsgAny, err := codectypes.NewAnyWithValue(migrateMsg)
		require.NoError(t, err)

		resp, err := msgSrv.MigrateAccount(f.app.Context(), &authtypes.MsgMigrateAccount{
			Signer:         f.mustAddr(addr),
			AccountType:    "base",
			AccountInitMsg: initMsgAny,
		})
		require.NoError(t, err)

		// check response semantics.
		require.Equal(t, resp.InitResponse.TypeUrl, "/cosmos.accounts.defaults.base.v1.MsgInitResponse")
		require.NotNil(t, resp.InitResponse.Value)

		// check the account was removed from x/auth and added to x/accounts
		require.Nil(t, f.authKeeper.GetAccount(f.app.Context(), addr))
		require.True(t, f.accountsKeeper.IsAccountsModuleAccount(f.app.Context(), addr))

		// check the init information is correctly propagated.
		seq, err := f.accountsKeeper.Query(f.app.Context(), addr, &basev1.QuerySequence{})
		require.NoError(t, err)
		require.Equal(t, migrateMsg.InitSequence, seq.(*basev1.QuerySequenceResponse).Sequence)

		pkResp, err := f.accountsKeeper.Query(f.app.Context(), addr, &basev1.QueryPubKey{})
		require.NoError(t, err)
		require.Equal(t, migrateMsg.PubKey, pkResp.(*basev1.QueryPubKeyResponse).PubKey)
	})
}
