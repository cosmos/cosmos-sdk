package types_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestDelegateAuthorizations(t *testing.T) {

	// verify MethodName
	delAuth, _ := stakingtypes.NewDelegateAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, &coin100)
	require.Equal(t, delAuth.MethodName(), "/cosmos.staking.v1beta1.Msg/Delegate")

	validators1_2 := []string{val1.String(), val2.String()}

	testCases := []struct {
		msg                  string
		allowed              []sdk.ValAddress
		denied               []sdk.ValAddress
		limit                *sdk.Coin
		srvMsg               sdk.ServiceMsg
		expectErr            bool
		isDelete             bool
		updatedAuthorization *stakingtypes.DelegateAuthorization
	}{
		{
			"expect 0 remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin100),
			false,
			true,
			nil,
		},
		{
			"verify remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin50),
			false,
			false,
			&stakingtypes.DelegateAuthorization{
				Validators: &stakingtypes.DelegateAuthorization_AllowList{
					AllowList: &stakingtypes.DelegateAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: &coin50},
		},
		{
			"testing with invalid validator",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val3, coin100),
			true,
			false,
			nil,
		},
		{
			"testing delegate without spent limit",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			nil,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val2, coin100),
			false,
			false,
			&stakingtypes.DelegateAuthorization{
				Validators: &stakingtypes.DelegateAuthorization_AllowList{
					AllowList: &stakingtypes.DelegateAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: nil},
		},
		{
			"fail: validator denied",
			[]sdk.ValAddress{},
			[]sdk.ValAddress{val1},
			nil,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin100),
			true,
			false,
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			delAuth, err := stakingtypes.NewDelegateAuthorization(tc.allowed, tc.denied, tc.limit)
			require.NoError(t, err)
			updated, del, err := delAuth.Accept(tc.srvMsg, tmproto.Header{})
			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, tc.isDelete, del)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.isDelete, del)
				if tc.updatedAuthorization != nil {
					require.Equal(t, tc.updatedAuthorization.String(), updated.String())
				}
			}
		})
	}
}

func createSrvMsgDelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
	msg := stakingtypes.NewMsgDelegate(delAddr, valAddr, amount)
	return sdk.ServiceMsg{
		MethodName: methodName,
		Request:    msg,
	}
}
