package types_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type AuthzTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func TestAuthzTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzTestSuite))
}

func (suite *AuthzTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	suite.ctx = ctx

}

var (
	coin100 = sdk.NewInt64Coin("steak", 100)
	coin50  = sdk.NewInt64Coin("steak", 50)
	delAddr = sdk.AccAddress("_____delegator _____")
	val1    = sdk.ValAddress("_____validator1_____")
	val2    = sdk.ValAddress("_____validator2_____")
	val3    = sdk.ValAddress("_____validator3_____")
)

func (suite *AuthzTestSuite) TestAuthzAuthorizations() {

	// verify MethodName
	delAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, &coin100)
	suite.Require().Equal(delAuth.MethodName(), stakingtypes.TypeDelegate)

	// error both allow & deny list
	_, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{val1}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, &coin100)
	suite.Require().Error(err)

	// verify MethodName
	undelAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, &coin100)
	suite.Require().Equal(undelAuth.MethodName(), stakingtypes.TypeUndelegate)

	// verify MethodName
	beginRedelAuth, _ := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{val1, val2}, []sdk.ValAddress{}, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE, &coin100)
	suite.Require().Equal(beginRedelAuth.MethodName(), stakingtypes.TypeBeginRedelegate)

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
		suite.Run(tc.msg, func() {
			delAuth, err := stakingtypes.NewStakeAuthorization(tc.allowed, tc.denied, tc.msgType, tc.limit)
			suite.Require().NoError(err)
			updated, del, err := delAuth.Accept(suite.ctx, tc.srvMsg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Equal(tc.isDelete, del)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.isDelete, del)
				if tc.updatedAuthorization != nil {
					suite.Require().Equal(tc.updatedAuthorization.String(), updated.String())
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
