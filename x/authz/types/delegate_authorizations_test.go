package types_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

var (
	coin100 = sdk.NewInt64Coin("steak", 100)
	coin50  = sdk.NewInt64Coin("steak", 50)
	delAddr = sdk.AccAddress("_____delegator _____")
	val1    = sdk.ValAddress("_____validator1_____")
	val2    = sdk.ValAddress("_____validator2_____")
	val3    = sdk.ValAddress("_____validator3_____")
)

func TestDelegateAuthorizations(t *testing.T) {
	delAuth := types.NewDelegateAuthorization([]sdk.ValAddress{val1, val2}, coin100)

	//verify MethodName
	require.Equal(t, delAuth.MethodName(), "/cosmos.staking.v1beta1.Msg/Delegate")

	// expect 0 remaining coins
	srvMsg := createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin100)
	updated, del, err := delAuth.Accept(srvMsg, tmproto.Header{})
	require.Equal(t, true, del)
	require.NoError(t, err)
	require.Nil(t, updated)

	// verify remaing coins
	delAuth = types.NewDelegateAuthorization([]sdk.ValAddress{val1, val2}, coin100)
	srvMsg = createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin50)
	updated, del, err = delAuth.Accept(srvMsg, tmproto.Header{})
	require.Equal(t, del, false)
	require.NoError(t, err)
	actual, ok := updated.(*types.DelegateAuthorization)
	require.True(t, ok)
	expected := types.DelegateAuthorization{ValidatorAddress: []string{val1.String(), val2.String()}, Amount: coin100.Sub(coin50)}
	require.Equal(t, expected.String(), actual.String())

	// fail over spent
	delAuth = types.NewDelegateAuthorization([]sdk.ValAddress{val1, val2}, coin100)
	srvMsg = createSrvMsgDelegate(delAuth.MethodName(), delAddr, val3, coin100.Add(coin50))
	updated, del, err = delAuth.Accept(srvMsg, tmproto.Header{})
	require.Error(t, err)
	require.Nil(t, updated)
	require.False(t, del)

	// fail with no validator
	delAuth = types.NewDelegateAuthorization([]sdk.ValAddress{val1, val2}, coin100)
	srvMsg = createSrvMsgDelegate(delAuth.MethodName(), delAddr, val3, coin50)
	updated, del, err = delAuth.Accept(srvMsg, tmproto.Header{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator not found")
	require.Nil(t, updated)
	require.False(t, del)

}

func createSrvMsgDelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
	msg := stakingtypes.NewMsgDelegate(delAddr, valAddr, amount)
	return sdk.ServiceMsg{
		MethodName: methodName,
		Request:    msg,
	}
}
