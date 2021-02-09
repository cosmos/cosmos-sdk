package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz/exported"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	_              authz.Authorization = &StakeAuthorization{}
	TypeDelegate                       = "/cosmos.staking.v1beta1.Msg/Delegate"
	TypeUndelegate                     = "/cosmos.staking.v1beta1.Msg/Undelegate"
)

// NewStakeAuthorization creates a new StakeAuthorization object.
func NewStakeAuthorization(allowed []sdk.ValAddress, denied []sdk.ValAddress, authzType string, amount *sdk.Coin) (*StakeAuthorization, error) {
	allowedValidators, deniedValidators, err := validateAndBech32fy(allowed, denied)
	if err != nil {
		return nil, err
	}

	authorization := StakeAuthorization{}
	if allowedValidators != nil {
		authorization.Validators = &StakeAuthorization_AllowList{AllowList: &StakeAuthorization_Validators{Address: allowedValidators}}
	} else {
		authorization.Validators = &StakeAuthorization_DenyList{DenyList: &StakeAuthorization_Validators{Address: deniedValidators}}
	}

	if amount != nil {
		authorization.MaxTokens = amount
	}
	authorization.AuthorizationType = authzType

	return &authorization, nil
}

// MethodName implements Authorization.MethodName.
func (authorization StakeAuthorization) MethodName() string {
	return authorization.AuthorizationType
}

// Accept implements Authorization.Accept.
func (authorization StakeAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated authz.Authorization, delete bool, err error) {
	var validatorAddress string
	var amount sdk.Coin

	switch msg := msg.Request.(type) {
	case *MsgDelegate:
		validatorAddress = msg.ValidatorAddress
		amount = msg.Amount
	case *MsgUndelegate:
		validatorAddress = msg.ValidatorAddress
		amount = msg.Amount
	default:
		return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unknown msg type")
	}

	isValidatorExists := false
	allowedList := authorization.GetAllowList().GetAddress()
	for _, validator := range allowedList {
		if validator == validatorAddress {
			isValidatorExists = true
			break
		}
	}
	denyList := authorization.GetDenyList().GetAddress()
	for _, validator := range denyList {
		if validator == validatorAddress {
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, " cannot delegate/undelegate to %s validator", validator)
		}
	}

	if !isValidatorExists {
		return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot delegate/undelegate")
	}

	if authorization.MaxTokens == nil {
		return &StakeAuthorization{Validators: authorization.GetValidators(), AuthorizationType: authorization.GetAuthorizationType()}, false, nil
	}

	limitLeft := authorization.MaxTokens.Sub(amount)
	if limitLeft.IsZero() {
		return nil, true, nil
	}

	return &StakeAuthorization{Validators: authorization.GetValidators(), MaxTokens: &limitLeft, AuthorizationType: authorization.GetAuthorizationType()}, false, nil

}
