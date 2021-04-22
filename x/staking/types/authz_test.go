package types_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	coin100 = sdk.NewInt64Coin("steak", 100)
	coin50  = sdk.NewInt64Coin("steak", 50)
	delAddr = sdk.AccAddress("_____delegator _____")
	val1    = sdk.ValAddress("_____validator1_____")
	val2    = sdk.ValAddress("_____validator2_____")
	val3    = sdk.ValAddress("_____validator3_____")
)

func TestAuthzAuthorizations(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// verify ValidateBasic returns error for the AUTHORIZATION_TYPE_UNSPECIFIED authorization type
	delAuth, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNSPECIFIED, &coin100)
	require.NoError(t, err)
	require.Error(t, delAuth.ValidateBasic())

	// verify MethodName
	delAuth, err = stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, &coin100)
	require.NoError(t, err)
	require.Equal(t, delAuth.MethodName(), stakingtypes.TypeDelegate)

	// error both allow & deny list
	_, err = stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{val1}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, &coin100)
	require.Error(t, err)

	// verify MethodName
	undelAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, &coin100)
	require.Equal(t, undelAuth.MethodName(), stakingtypes.TypeUndelegate)

	// verify MethodName
	beginRedelAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE, &coin100)
	require.Equal(t, beginRedelAuth.MethodName(), stakingtypes.TypeBeginRedelegate)

	validators1_2 := []string{val1.String(), val2.String()}

	testCases := []struct {
		msg                  string
		allowed              []sdk.ValAddress
		denied               []sdk.ValAddress
		msgType              stakingtypes.AuthorizationType
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
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
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
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
			&coin100,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val1, coin50),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: &coin50, AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE},
		},
		{
			"delegate: testing with invalid validator",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
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
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
			nil,
			createSrvMsgDelegate(delAuth.MethodName(), delAddr, val2, coin100),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: nil, AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE},
		},
		{
			"delegate: fail validator denied",
			[]sdk.ValAddress{},
			[]sdk.ValAddress{val1},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
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
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE,
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
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE,
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val1, coin50),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: &coin50, AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE},
		},
		{
			"undelegate: testing with invalid validator",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE,
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
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE,
			nil,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val2, coin100),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: nil, AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE},
		},
		{
			"undelegate: fail cannot undelegate, permission denied",
			[]sdk.ValAddress{},
			[]sdk.ValAddress{val1},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE,
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val1, coin100),
			true,
			false,
			nil,
		},

		{
			"redelegate: expect 0 remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE,
			&coin100,
			createSrvMsgUndelegate(undelAuth.MethodName(), delAddr, val1, coin100),
			false,
			true,
			nil,
		},
		{
			"redelegate: verify remaining coins",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE,
			&coin100,
			createSrvMsgReDelegate(undelAuth.MethodName(), delAddr, val1, coin50),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: &coin50, AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE},
		},
		{
			"redelegate: testing with invalid validator",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE,
			&coin100,
			createSrvMsgReDelegate(undelAuth.MethodName(), delAddr, val3, coin100),
			true,
			false,
			nil,
		},
		{
			"redelegate: testing delegate without spent limit",
			[]sdk.ValAddress{val1, val2},
			[]sdk.ValAddress{},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE,
			nil,
			createSrvMsgReDelegate(undelAuth.MethodName(), delAddr, val2, coin100),
			false,
			false,
			&stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: validators1_2},
				}, MaxTokens: nil, AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE},
		},
		{
			"redelegate: fail cannot undelegate, permission denied",
			[]sdk.ValAddress{},
			[]sdk.ValAddress{val1},
			stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE,
			&coin100,
			createSrvMsgReDelegate(undelAuth.MethodName(), delAddr, val1, coin100),
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
			resp, err := delAuth.Accept(ctx, tc.srvMsg)
			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, tc.isDelete, del)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.isDelete, resp.Delete)
				if tc.updatedAuthorization != nil {
					require.Equal(t, tc.updatedAuthorization.String(), resp.Updated.String())
				}
			}
		})
	}
}

func createSrvMsgUndelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
	msg := stakingtypes.NewMsgUndelegate(delAddr, valAddr, amount)
	return sdk.ServiceMsg{
		MethodName: methodName,
		Request:    msg,
	}
}

func createSrvMsgReDelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
	msg := stakingtypes.NewMsgBeginRedelegate(delAddr, valAddr, valAddr, amount)
	return sdk.ServiceMsg{
		MethodName: methodName,
		Request:    msg,
	}
}

func createSrvMsgDelegate(methodName string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) sdk.ServiceMsg {
	msg := stakingtypes.NewMsgDelegate(delAddr, valAddr, amount)
	return sdk.ServiceMsg{
		MethodName: methodName,
		Request:    msg,
	}
}
