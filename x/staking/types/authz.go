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

	TypeDelegate        = "/cosmos.staking.v1beta1.Msg/Delegate"
	TypeUndelegate      = "/cosmos.staking.v1beta1.Msg/Undelegate"
	TypeBeginRedelegate = "/cosmos.staking.v1beta1.Msg/BeginRedelegate"
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

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a StakeAuthorization) MsgTypeURL() string {
	authzType, err := normalizeAuthzType(a.AuthorizationType)
	if err != nil {
		panic(err)
	}
	return authzType
}

func (a StakeAuthorization) ValidateBasic() error {
	if a.MaxTokens != nil && a.MaxTokens.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "negative coin amount: %v", a.MaxTokens)
	}
	if a.AuthorizationType == AuthorizationType_AUTHORIZATION_TYPE_UNSPECIFIED {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "unknown authorization type")
	}

	return nil
}

// Accept implements Authorization.Accept.
// func (authorization StakeAuthorization) Accept(ctx sdk.Context, msg sdk.ServiceMsg) (updated authz.Authorization, delete bool, err error) {
func (a StakeAuthorization) Accept(ctx sdk.Context, msg sdk.ServiceMsg) (authz.AcceptResponse, error) {
	var validatorAddress string
	var amount sdk.Coin

	switch msg := msg.Request.(type) {
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
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidRequest.Wrap("unknown msg type")
	}

	isValidatorExists := false
	allowedList := a.GetAllowList().GetAddress()
	for _, validator := range allowedList {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "stake authorization")
		if validator == validatorAddress {
			isValidatorExists = true
			break
		}
	}

	denyList := a.GetDenyList().GetAddress()
	for _, validator := range denyList {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "stake authorization")
		if validator == validatorAddress {
			return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf(" cannot delegate/undelegate to %s validator", validator)
		}
	}

	if !isValidatorExists {
		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("cannot delegate/undelegate to %s validator", validatorAddress)
	}

	if a.MaxTokens == nil {
		return authz.AcceptResponse{Accept: true, Delete: false,
			Updated: &StakeAuthorization{Validators: a.GetValidators(), AuthorizationType: a.GetAuthorizationType()}}, nil
	}

	limitLeft := a.MaxTokens.Sub(amount)
	if limitLeft.IsZero() {
		return authz.AcceptResponse{Accept: true, Delete: true}, nil
	}
	return authz.AcceptResponse{Accept: true, Delete: false,
		Updated: &StakeAuthorization{Validators: a.GetValidators(), AuthorizationType: a.GetAuthorizationType(), MaxTokens: &limitLeft}}, nil
}

func validateAndBech32fy(allowed []sdk.ValAddress, denied []sdk.ValAddress) ([]string, []string, error) {
	if len(allowed) == 0 && len(denied) == 0 {
		return nil, nil, sdkerrors.ErrInvalidRequest.Wrap("both allowed & deny list cannot be empty")
	}

	if len(allowed) > 0 && len(denied) > 0 {
		return nil, nil, sdkerrors.ErrInvalidRequest.Wrap("cannot set both allowed & deny list")
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
		return TypeDelegate, nil
	case AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE:
		return TypeUndelegate, nil
	case AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE:
		return TypeBeginRedelegate, nil
	default:
		return "", sdkerrors.ErrInvalidType.Wrapf("unknown authorization type %T", authzType)
	}
}
