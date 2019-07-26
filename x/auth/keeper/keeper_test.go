package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestAccountMapperGetSet(t *testing.T) {
	input := SetupTestInput()
	addr := sdk.AccAddress([]byte("some-address"))

	// no account before its created
	acc := input.AccountKeeper.GetAccount(input.Ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())
	require.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, input.AccountKeeper.GetAccount(input.Ctx, addr))

	// set some values on the account and save it
	newSequence := uint64(20)
	acc.SetSequence(newSequence)
	input.AccountKeeper.SetAccount(input.Ctx, acc)

	// check the new values
	acc = input.AccountKeeper.GetAccount(input.Ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, newSequence, acc.GetSequence())
}

func TestAccountMapperRemoveAccount(t *testing.T) {
	input := SetupTestInput()
	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	// create accounts
	acc1 := input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr1)
	acc2 := input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr2)

	accSeq1 := uint64(20)
	accSeq2 := uint64(40)

	acc1.SetSequence(accSeq1)
	acc2.SetSequence(accSeq2)
	input.AccountKeeper.SetAccount(input.Ctx, acc1)
	input.AccountKeeper.SetAccount(input.Ctx, acc2)

	acc1 = input.AccountKeeper.GetAccount(input.Ctx, addr1)
	require.NotNil(t, acc1)
	require.Equal(t, accSeq1, acc1.GetSequence())

	// remove one account
	input.AccountKeeper.RemoveAccount(input.Ctx, acc1)
	acc1 = input.AccountKeeper.GetAccount(input.Ctx, addr1)
	require.Nil(t, acc1)

	acc2 = input.AccountKeeper.GetAccount(input.Ctx, addr2)
	require.NotNil(t, acc2)
	require.Equal(t, accSeq2, acc2.GetSequence())
}

func TestSetParams(t *testing.T) {
	input := SetupTestInput()
	params := types.DefaultParams()

	input.AccountKeeper.SetParams(input.Ctx, params)

	newParams := types.Params{}
	input.AccountKeeper.paramSubspace.Get(input.Ctx, types.KeyTxSigLimit, &newParams.TxSigLimit)
	require.Equal(t, newParams.TxSigLimit, types.DefaultTxSigLimit)
}

func TestGetParams(t *testing.T) {
	input := SetupTestInput()
	params := types.DefaultParams()

	input.AccountKeeper.SetParams(input.Ctx, params)

	newParams := input.AccountKeeper.GetParams(input.Ctx)
	require.Equal(t, params, newParams)
}
