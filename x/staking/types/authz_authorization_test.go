package types_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestAuthzAuthorizations(t *testing.T) {

	// verify MethodName
	delAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, "/cosmos.staking.v1beta1.Msg/Delegate", &coin100)
	require.Equal(t, delAuth.MethodName(), "/cosmos.staking.v1beta1.Msg/Delegate")

	// verify MethodName
	undelAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, "/cosmos.staking.v1beta1.Msg/Undelegate", &coin100)
	require.Equal(t, undelAuth.MethodName(), "/cosmos.staking.v1beta1.Msg/Undelegate")

	validators1_2 := []string{val1.String(), val2.String()}

	testCases := []struct {
		msg                  string
		allowed              []sdk.ValAddress
		denied               []sdk.ValAddress
		msgType              string
		limit                *sdk.Coin
		srvMsg               sdk.ServiceMsg
		expectErr            bool
		isDelete             bool
		updatedAuthorization *stakingtypes.StakeAuthorization
	}{
		{
			"delegate: expect 0 remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Delegate",
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin100),
			false,
			true,
			nil,
		},
		{
			"delegate: verify remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Delegate",
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin50),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: &coin50, AuthorizationType: stakingtypes.TypeDelegate},
		},
		{
			"delegate: testing with invalid validator",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Delegate",
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val3, coin100),
			true,
			false,
			nil,
		},
		{
			"delegate: testing delegate without spent limit",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Delegate",
			nil,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val2, coin100),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: nil, AuthorizationType: stakingtypes.TypeDelegate},
		},
		{
			"delegate: fail validator denied",
			[]sdk.ValAddress{},
			[]sdk.ValAddress{val1},
			"/cosmos.staking.v1beta1.Msg/Delegate",
			nil,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin100),
			true,
			false,
			nil,
		},

		{
			"undelegate: expect 0 remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/undelegate",
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val1, coin100),
			false,
			true,
			nil,
		},
		{
			"undelegate: verify remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Undelegate",
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val1, coin50),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: &coin50, AuthorizationType: stakingtypes.TypeUndelegate},
		},
		{
			"undelegate: testing with invalid validator",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Undelegate",
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val3, coin100),
			true,
			false,
			nil,
		},
		{
			"undelegate: testing delegate without spent limit",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			"/cosmos.staking.v1beta1.Msg/Undelegate",
			nil,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val2, coin100),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: nil, AuthorizationType: stakingtypes.TypeUndelegate},
		},
		{
			"undelegate: fail cannot undelegate, permission denied",
			[]sdk.ValAddress{},
			[]sdk.ValAddress{val1},
			"/cosmos.staking.v1beta1.Msg/Undelegate",
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val1, coin100),
			true,
			false,
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			delAuth, err := stakingtypes.NewStakeAuthorization(tc.allowed, tc.denied, tc.msgType, tc.limit)
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

// func createSrvMsgDelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
// 	msg := stakingtypes.NewMsgDelegate(delAddr, valAddr, amount)
// 	return sdk.ServiceMsg{
// 		MethodName: methodName,
// 		Request:    msg,
// 	}
// }

func createSrvMsgUnDelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
	msg := stakingtypes.NewMsgUndelegate(delAddr, valAddr, amount)
	return sdk.ServiceMsg{
		MethodName: methodName,
		Request:    msg,
	}
}
