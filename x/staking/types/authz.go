package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz/exported"
)

// TODO: Revisit this once we have propoer gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054, https://github.com/cosmos/cosmos-sdk/discussions/9072
const gasCostPerIteration = uint64(10)

var (
	_ authz.Authorization = &StakeAuthorization{}
)

// NewStakeAuthorization creates a new StakeAuthorization object.
func NewStakeAuthorization(allowed []sdk.ValAddress, denied []sdk.ValAddress, authzType AuthorizationType, amount *sdk.Coin) (*StakeAuthorization, error) {
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
	authzType, err := normalizeAuthzType(authorization.AuthorizationType)
	if err != nil {
		panic(err)
	}
	return authzType
}

func (authorization StakeAuthorization) ValidateBasic() error {
	if authorization.MaxTokens != nil && authorization.MaxTokens.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "negative coin amount: %v", authorization.MaxTokens)
	}
	if authorization.AuthorizationType == AuthorizationType_AUTHORIZATION_TYPE_UNSPECIFIED {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "unknown authorization type")
	}

	return nil
}

// Accept implements Authorization.Accept.
func (authorization StakeAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (updated authz.Authorization, delete bool, err error) {
	var validatorAddress string
	var amount sdk.Coin

	switch msg := msg.(type) {
	case *MsgDelegate:
		validatorAddress = msg.ValidatorAddress
		amount = msg.Amount
	case *MsgUndelegate:
		validatorAddress = msg.ValidatorAddress
		amount = msg.Amount
	case *MsgBeginRedelegate:
		validatorAddress = msg.ValidatorDstAddress
		amount = msg.Amount
	default:
		return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unknown msg type")
	}

	isValidatorExists := false
	allowedList := authorization.GetAllowList().GetAddress()
	for _, validator := range allowedList {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "stake authorization")
		if validator == validatorAddress {
			isValidatorExists = true
			break
		}
	}

	denyList := authorization.GetDenyList().GetAddress()
	for _, validator := range denyList {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "stake authorization")
		if validator == validatorAddress {
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, " cannot delegate/undelegate to %s validator", validator)
		}
	}

	if !isValidatorExists {
		return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot delegate/undelegate to %s validator", validatorAddress)
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

func validateAndBech32fy(allowed []sdk.ValAddress, denied []sdk.ValAddress) ([]string, []string, error) {
	if len(allowed) == 0 && len(denied) == 0 {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "both allowed & deny list cannot be empty")
	}

	if len(allowed) > 0 && len(denied) > 0 {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot set both allowed & deny list")
	}

	allowedValidators := make([]string, len(allowed))
	if len(allowed) > 0 {
		for i, validator := range allowed {
			allowedValidators[i] = validator.String()
		}
		return allowedValidators, nil, nil
	}

	deniedValidators := make([]string, len(denied))
	for i, validator := range denied {
		deniedValidators[i] = validator.String()
	}

	return nil, deniedValidators, nil
}

func normalizeAuthzType(authzType AuthorizationType) (string, error) {
	switch authzType {
	case AuthorizationType_AUTHORIZATION_TYPE_DELEGATE:
		return sdk.MsgTypeURL(&MsgDelegate{}), nil
	case AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE:
		return sdk.MsgTypeURL(&MsgUndelegate{}), nil
	case AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE:
		return sdk.MsgTypeURL(&MsgBeginRedelegate{}), nil
	default:
		return "", sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "unknown authorization type %T", authzType)
	}
}
