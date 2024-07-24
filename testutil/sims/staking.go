package sims

import (
	"fmt"
	"time"

	gogoany "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StakingNotBondedPoolName = "not_bonded_tokens_pool"
	StakingBondedPoolName    = "bonded_tokens_pool"
)

// LastValidatorPower required for validator set update logic.
type LastValidatorPower struct {
	// address is the address of the validator.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// power defines the power of the validator.
	Power int64 `protobuf:"varint,2,opt,name=power,proto3" json:"power,omitempty"`
}

type StakingGenesisState struct {
	// params defines all the parameters of related to deposit.
	Params StakingParams `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// last_total_power tracks the total amounts of bonded tokens recorded during
	// the previous end block.
	LastTotalPower math.Int `protobuf:"bytes,2,opt,name=last_total_power,json=lastTotalPower,proto3,customtype=cosmossdk.io/math.Int" json:"last_total_power"`
	// last_validator_powers is a special index that provides a historical list
	// of the last-block's bonded validators.
	LastValidatorPowers []LastValidatorPower `protobuf:"bytes,3,rep,name=last_validator_powers,json=lastValidatorPowers,proto3" json:"last_validator_powers"`
	// validators defines the validator set at genesis.
	Validators []StakingValidator `protobuf:"bytes,4,rep,name=validators,proto3" json:"validators"`
	// delegations defines the delegations active at genesis.
	Delegations []StakingDelegation `protobuf:"bytes,5,rep,name=delegations,proto3" json:"delegations"`
	// exported defines a bool to identify whether the chain dealing with exported or initialized genesis.
	Exported bool `protobuf:"varint,8,opt,name=exported,proto3" json:"exported,omitempty"`
}

func (s *StakingGenesisState) ToProto() (*anypb.Any, error) {
	params, err := s.Params.ToProto()
	if err != nil {
		return nil, err
	}

	lastValidatorPowers := make([]map[string]interface{}, len(s.LastValidatorPowers))
	for i, lvp := range s.LastValidatorPowers {
		lastValidatorPowers[i] = map[string]interface{}{
			"address": lvp.Address,
			"power":   lvp.Power,
		}
	}

	validators := make([]map[string]interface{}, len(s.Validators))
	for i, v := range s.Validators {
		validatorProto, err := v.ToProto()
		if err != nil {
			return nil, err
		}
		validators[i] = map[string]interface{}{"validator": validatorProto}
	}

	delegations := make([]map[string]interface{}, len(s.Delegations))
	for i, d := range s.Delegations {
		delegationProto, err := d.ToProto()
		if err != nil {
			return nil, err
		}
		delegations[i] = map[string]interface{}{"delegation": delegationProto}
	}

	fields := map[string]interface{}{
		"params":                params,
		"last_total_power":      s.LastTotalPower.String(),
		"last_validator_powers": lastValidatorPowers,
		"validators":            validators,
		"delegations":           delegations,
		"exported":              s.Exported,
	}

	pbStruct, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}

	return anypb.New(pbStruct)
}

func ProtoToStakingGenesisState(protoMsg *anypb.Any) (*StakingGenesisState, error) {
	var s structpb.Struct
	if err := protoMsg.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	genesisState := &StakingGenesisState{}

	// Parse Params
	paramsAny, err := anypb.New(s.Fields["params"].GetStructValue())
	if err != nil {
		return nil, err
	}
	params, err := ProtoToStakingParams(paramsAny)
	if err != nil {
		return nil, err
	}
	genesisState.Params = *params

	// Parse LastTotalPower
	lastTotalPower, ok := math.NewIntFromString(s.Fields["last_total_power"].GetStringValue())
	if !ok {
		return nil, fmt.Errorf("failed to parse last_total_power")
	}
	genesisState.LastTotalPower = lastTotalPower

	// Parse LastValidatorPowers
	lastValidatorPowersValue := s.Fields["last_validator_powers"].GetListValue()
	genesisState.LastValidatorPowers = make([]LastValidatorPower, len(lastValidatorPowersValue.Values))
	for i, v := range lastValidatorPowersValue.Values {
		lvpStruct := v.GetStructValue()
		genesisState.LastValidatorPowers[i] = LastValidatorPower{
			Address: lvpStruct.Fields["address"].GetStringValue(),
			Power:   int64(lvpStruct.Fields["power"].GetNumberValue()),
		}
	}

	// Parse Validators
	validatorsValue := s.Fields["validators"].GetListValue()
	genesisState.Validators = make([]StakingValidator, len(validatorsValue.Values))
	for i, v := range validatorsValue.Values {
		validatorAny, err := anypb.New(v.GetStructValue().Fields["validator"].GetStructValue())
		if err != nil {
			return nil, err
		}
		validator, err := ProtoToStakingValidator(validatorAny)
		if err != nil {
			return nil, err
		}
		genesisState.Validators[i] = *validator
	}

	// Parse Delegations
	delegationsValue := s.Fields["delegations"].GetListValue()
	genesisState.Delegations = make([]StakingDelegation, len(delegationsValue.Values))
	for i, v := range delegationsValue.Values {
		delegationAny, err := anypb.New(v.GetStructValue().Fields["delegation"].GetStructValue())
		if err != nil {
			return nil, err
		}
		delegation, err := ProtoToStakingDelegation(delegationAny)
		if err != nil {
			return nil, err
		}
		genesisState.Delegations[i] = *delegation
	}

	// Parse Exported
	genesisState.Exported = s.Fields["exported"].GetBoolValue()

	return genesisState, nil
}

type StakingParams struct {
	// unbonding_time is the time duration of unbonding.
	UnbondingTime time.Duration `protobuf:"bytes,1,opt,name=unbonding_time,json=unbondingTime,proto3,stdduration" json:"unbonding_time"`
	// max_validators is the maximum number of validators.
	MaxValidators uint32 `protobuf:"varint,2,opt,name=max_validators,json=maxValidators,proto3" json:"max_validators,omitempty"`
	// max_entries is the max entries for either unbonding delegation or redelegation (per pair/trio).
	MaxEntries uint32 `protobuf:"varint,3,opt,name=max_entries,json=maxEntries,proto3" json:"max_entries,omitempty"`
	// historical_entries is the number of historical entries to persist.
	HistoricalEntries uint32 `protobuf:"varint,4,opt,name=historical_entries,json=historicalEntries,proto3" json:"historical_entries,omitempty"`
	// bond_denom defines the bondable coin denomination.
	BondDenom string `protobuf:"bytes,5,opt,name=bond_denom,json=bondDenom,proto3" json:"bond_denom,omitempty"`
	// min_commission_rate is the chain-wide minimum commission rate that a validator can charge their delegators
	MinCommissionRate math.LegacyDec `protobuf:"bytes,6,opt,name=min_commission_rate,json=minCommissionRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"min_commission_rate" yaml:"min_commission_rate"`
	// key_rotation_fee is fee to be spent when rotating validator's key
	// (either consensus pubkey or operator key)
	KeyRotationFee sdk.Coin `protobuf:"bytes,7,opt,name=key_rotation_fee,json=keyRotationFee,proto3" json:"key_rotation_fee"`
}

func (s *StakingParams) ToProto() (*anypb.Any, error) {
	fields := map[string]interface{}{
		"unbonding_time":      s.UnbondingTime.String(),
		"max_validators":      s.MaxValidators,
		"max_entries":         s.MaxEntries,
		"historical_entries":  s.HistoricalEntries,
		"bond_denom":          s.BondDenom,
		"min_commission_rate": s.MinCommissionRate.String(),
		"key_rotation_fee": map[string]interface{}{
			"denom":  s.KeyRotationFee.Denom,
			"amount": s.KeyRotationFee.Amount.String(),
		},
	}

	pbStruct, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}

	return anypb.New(pbStruct)
}

func ProtoToStakingParams(protoMsg *anypb.Any) (*StakingParams, error) {
	var s structpb.Struct
	if err := protoMsg.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	params := &StakingParams{}

	unbondingTime, err := time.ParseDuration(s.Fields["unbonding_time"].GetStringValue())
	if err != nil {
		return nil, err
	}
	params.UnbondingTime = unbondingTime

	params.MaxValidators = uint32(s.Fields["max_validators"].GetNumberValue())
	params.MaxEntries = uint32(s.Fields["max_entries"].GetNumberValue())
	params.HistoricalEntries = uint32(s.Fields["historical_entries"].GetNumberValue())
	params.BondDenom = s.Fields["bond_denom"].GetStringValue()

	minCommissionRate, err := math.LegacyNewDecFromStr(s.Fields["min_commission_rate"].GetStringValue())
	if err != nil {
		return nil, err
	}
	params.MinCommissionRate = minCommissionRate

	keyRotationFeeStruct := s.Fields["key_rotation_fee"].GetStructValue()
	amount, ok := math.NewIntFromString(keyRotationFeeStruct.Fields["amount"].GetStringValue())
	if !ok {
		return nil, fmt.Errorf("failed to parse key_rotation_fee amount")
	}
	params.KeyRotationFee = sdk.NewCoin(
		keyRotationFeeStruct.Fields["denom"].GetStringValue(),
		amount,
	)

	return params, nil
}

type StakingValidator struct {
	// operator_address defines the address of the validator's operator; bech encoded in JSON.
	OperatorAddress string `protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	// consensus_pubkey is the consensus public key of the validator, as a Protobuf Any.
	ConsensusPubkey *gogoany.Any `protobuf:"bytes,2,opt,name=consensus_pubkey,json=consensusPubkey,proto3" json:"consensus_pubkey,omitempty"`
	// jailed defined whether the validator has been jailed from bonded status or not.
	Jailed bool `protobuf:"varint,3,opt,name=jailed,proto3" json:"jailed,omitempty"`
	// status is the validator status (bonded/unbonding/unbonded).
	Status int32 `protobuf:"varint,4,opt,name=status,proto3" json:"status,omitempty"`
	// tokens define the delegated tokens (incl. self-delegation).
	Tokens math.Int `protobuf:"bytes,5,opt,name=tokens,proto3,customtype=cosmossdk.io/math.Int" json:"tokens"`
	// delegator_shares defines total shares issued to a validator's delegators.
	DelegatorShares math.LegacyDec `protobuf:"bytes,6,opt,name=delegator_shares,json=delegatorShares,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"delegator_shares"`
	// description defines the description terms for the validator.
	Description StakingDescription `protobuf:"bytes,7,opt,name=description,proto3" json:"description"`
	// unbonding_height defines, if unbonding, the height at which this validator has begun unbonding.
	UnbondingHeight int64 `protobuf:"varint,8,opt,name=unbonding_height,json=unbondingHeight,proto3" json:"unbonding_height,omitempty"`
	// unbonding_time defines, if unbonding, the min time for the validator to complete unbonding.
	UnbondingTime time.Time `protobuf:"bytes,9,opt,name=unbonding_time,json=unbondingTime,proto3,stdtime" json:"unbonding_time"`
	// commission defines the commission parameters.
	Commission StakingValidatorCommission `protobuf:"bytes,10,opt,name=commission,proto3" json:"commission"`
	// min_self_delegation is the validator's self declared minimum self delegation.
	MinSelfDelegation math.Int `protobuf:"bytes,11,opt,name=min_self_delegation,json=minSelfDelegation,proto3,customtype=cosmossdk.io/math.Int" json:"min_self_delegation"`
	// strictly positive if this validator's unbonding has been stopped by external modules
	UnbondingOnHoldRefCount int64 `protobuf:"varint,12,opt,name=unbonding_on_hold_ref_count,json=unbondingOnHoldRefCount,proto3" json:"unbonding_on_hold_ref_count,omitempty"`
	// list of unbonding ids, each uniquely identifying an unbonding of this validator
	UnbondingIds []uint64 `protobuf:"varint,13,rep,packed,name=unbonding_ids,json=unbondingIds,proto3" json:"unbonding_ids,omitempty"`
}

func (s *StakingValidator) ToProto() (*anypb.Any, error) {
	consensusPubkeyAny, err := gogoany.NewAnyWithCacheWithValue(s.ConsensusPubkey)
	if err != nil {
		return nil, err
	}

	descriptionProto, err := s.Description.ToProto()
	if err != nil {
		return nil, err
	}

	commissionProto, err := s.Commission.ToProto()
	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{
		"operator_address":            s.OperatorAddress,
		"consensus_pubkey":            consensusPubkeyAny,
		"jailed":                      s.Jailed,
		"status":                      s.Status,
		"tokens":                      s.Tokens.String(),
		"delegator_shares":            s.DelegatorShares.String(),
		"description":                 descriptionProto,
		"unbonding_height":            s.UnbondingHeight,
		"unbonding_time":              s.UnbondingTime.Format(time.RFC3339),
		"commission":                  commissionProto,
		"min_self_delegation":         s.MinSelfDelegation.String(),
		"unbonding_on_hold_ref_count": s.UnbondingOnHoldRefCount,
		"unbonding_ids":               s.UnbondingIds,
	}

	pbStruct, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}

	return anypb.New(pbStruct)
}

func ProtoToStakingValidator(protoMsg *anypb.Any) (*StakingValidator, error) {
	var s structpb.Struct
	if err := protoMsg.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	validator := &StakingValidator{}

	validator.OperatorAddress = s.Fields["operator_address"].GetStringValue()

	consensusPubkeyAny := s.Fields["consensus_pubkey"].GetStructValue()
	validator.ConsensusPubkey = &gogoany.Any{
		TypeUrl: consensusPubkeyAny.Fields["type_url"].GetStringValue(),
		Value:   []byte(consensusPubkeyAny.Fields["value"].GetStringValue()),
	}

	validator.Jailed = s.Fields["jailed"].GetBoolValue()
	validator.Status = int32(s.Fields["status"].GetNumberValue())

	tokens, ok := math.NewIntFromString(s.Fields["tokens"].GetStringValue())
	if !ok {
		return nil, fmt.Errorf("failed to parse tokens")
	}
	validator.Tokens = tokens

	delegatorShares, err := math.LegacyNewDecFromStr(s.Fields["delegator_shares"].GetStringValue())
	if err != nil {
		return nil, err
	}
	validator.DelegatorShares = delegatorShares

	descriptionAny, err := anypb.New(s.Fields["description"].GetStructValue())
	if err != nil {
		return nil, err
	}
	description, err := ProtoToStakingDescription(descriptionAny)
	if err != nil {
		return nil, err
	}
	validator.Description = *description

	validator.UnbondingHeight = int64(s.Fields["unbonding_height"].GetNumberValue())

	unbondingTime, err := time.Parse(time.RFC3339, s.Fields["unbonding_time"].GetStringValue())
	if err != nil {
		return nil, err
	}
	validator.UnbondingTime = unbondingTime

	commissionAny, err := anypb.New(s.Fields["commission"].GetStructValue())
	if err != nil {
		return nil, err
	}
	commission, err := ProtoToStakingValidatorCommission(commissionAny)
	if err != nil {
		return nil, err
	}
	validator.Commission = *commission

	minSelfDelegation, ok := math.NewIntFromString(s.Fields["min_self_delegation"].GetStringValue())
	if !ok {
		return nil, fmt.Errorf("failed to parse min_self_delegation")
	}
	validator.MinSelfDelegation = minSelfDelegation

	validator.UnbondingOnHoldRefCount = int64(s.Fields["unbonding_on_hold_ref_count"].GetNumberValue())

	unbondingIdsValue := s.Fields["unbonding_ids"].GetListValue()
	validator.UnbondingIds = make([]uint64, len(unbondingIdsValue.Values))
	for i, v := range unbondingIdsValue.Values {
		validator.UnbondingIds[i] = uint64(v.GetNumberValue())
	}

	return validator, nil
}

// StakingDescription defines a validator description.
type StakingDescription struct {
	// moniker defines a human-readable name for the validator.
	Moniker string `protobuf:"bytes,1,opt,name=moniker,proto3" json:"moniker,omitempty"`
	// identity defines an optional identity signature (ex. UPort or Keybase).
	Identity string `protobuf:"bytes,2,opt,name=identity,proto3" json:"identity,omitempty"`
	// website defines an optional website link.
	Website string `protobuf:"bytes,3,opt,name=website,proto3" json:"website,omitempty"`
	// security_contact defines an optional email for security contact.
	SecurityContact string `protobuf:"bytes,4,opt,name=security_contact,json=securityContact,proto3" json:"security_contact,omitempty"`
	// details define other optional details.
	Details string `protobuf:"bytes,5,opt,name=details,proto3" json:"details,omitempty"`
}

func (s *StakingDescription) ToProto() (*anypb.Any, error) {
	fields := map[string]interface{}{
		"moniker":          s.Moniker,
		"identity":         s.Identity,
		"website":          s.Website,
		"security_contact": s.SecurityContact,
		"details":          s.Details,
	}

	pbStruct, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}

	return anypb.New(pbStruct)
}

func ProtoToStakingDescription(protoMsg *anypb.Any) (*StakingDescription, error) {
	var s structpb.Struct
	if err := protoMsg.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	return &StakingDescription{
		Moniker:         s.Fields["moniker"].GetStringValue(),
		Identity:        s.Fields["identity"].GetStringValue(),
		Website:         s.Fields["website"].GetStringValue(),
		SecurityContact: s.Fields["security_contact"].GetStringValue(),
		Details:         s.Fields["details"].GetStringValue(),
	}, nil
}

// StakingDelegation defines the structure for delegated funds per delegator.
type StakingDelegation struct {
	// delegator_address is the encoded address of the delegator.
	DelegatorAddress string `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	// validator_address is the encoded address of the validator.
	ValidatorAddress string `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	// shares define the delegation shares received.
	Shares math.LegacyDec `protobuf:"bytes,3,opt,name=shares,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"shares"`
}

func (s *StakingDelegation) ToProto() (*anypb.Any, error) {
	// Create a map to hold the protobuf fields
	fields := map[string]interface{}{
		"delegator_address": s.DelegatorAddress,
		"validator_address": s.ValidatorAddress,
		"shares":            s.Shares.String(),
	}

	// Convert the map to a protobuf Struct
	pbStruct, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}

	return anypb.New(pbStruct)
}

func ProtoToStakingDelegation(protoMsg *anypb.Any) (*StakingDelegation, error) {
	var s structpb.Struct
	if err := protoMsg.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	shares, err := math.LegacyNewDecFromStr(s.Fields["shares"].GetStringValue())
	if err != nil {
		return nil, err
	}

	return &StakingDelegation{
		DelegatorAddress: s.Fields["delegator_address"].GetStringValue(),
		ValidatorAddress: s.Fields["validator_address"].GetStringValue(),
		Shares:           shares,
	}, nil
}

// CommissionRates defines the initial commission rates to be used for creating a validator.
type StakingCommissionRates struct {
	// rate is the commission rate charged to delegators, as a fraction.
	Rate math.LegacyDec `protobuf:"bytes,1,opt,name=rate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"rate"`
	// max_rate defines the maximum commission rate which validator can ever charge, as a fraction.
	MaxRate math.LegacyDec `protobuf:"bytes,2,opt,name=max_rate,json=maxRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"max_rate"`
	// max_change_rate defines the maximum daily increase of the validator commission, as a fraction.
	MaxChangeRate math.LegacyDec `protobuf:"bytes,3,opt,name=max_change_rate,json=maxChangeRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"max_change_rate"`
}

// StakingValidatorCommission defines the initial commission rates to be used for creating a validator.
type StakingValidatorCommission struct {
	// commission_rates defines the initial commission rates to be used for creating a validator.
	StakingCommissionRates `protobuf:"bytes,1,opt,name=commission_rates,json=commissionRates,proto3,embedded=commission_rates" json:"commission_rates"`
	// update_time is the last time the commission rate was changed.
	UpdateTime time.Time `protobuf:"bytes,2,opt,name=update_time,json=updateTime,proto3,stdtime" json:"update_time"`
}

func (s *StakingValidatorCommission) ToProto() (*anypb.Any, error) {
	// Create a map to hold the protobuf fields
	fields := map[string]interface{}{
		"rate":            s.Rate.String(),
		"max_rate":        s.MaxRate.String(),
		"max_change_rate": s.MaxChangeRate.String(),
	}

	// Convert the map to a protobuf Struct
	pbStruct, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}

	return anypb.New(pbStruct)
}

func ProtoToStakingValidatorCommission(protoMsg *anypb.Any) (*StakingValidatorCommission, error) {
	var s structpb.Struct
	if err := protoMsg.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	commission := &StakingValidatorCommission{}

	commissionRates := s.Fields["commission_rates"].GetStructValue()
	rate, err := math.LegacyNewDecFromStr(commissionRates.Fields["rate"].GetStringValue())
	if err != nil {
		return nil, err
	}
	maxRate, err := math.LegacyNewDecFromStr(commissionRates.Fields["max_rate"].GetStringValue())
	if err != nil {
		return nil, err
	}
	maxChangeRate, err := math.LegacyNewDecFromStr(commissionRates.Fields["max_change_rate"].GetStringValue())
	if err != nil {
		return nil, err
	}

	commission.StakingCommissionRates = StakingCommissionRates{
		Rate:          rate,
		MaxRate:       maxRate,
		MaxChangeRate: maxChangeRate,
	}

	updateTime, err := time.Parse(time.RFC3339, s.Fields["update_time"].GetStringValue())
	if err != nil {
		return nil, err
	}
	commission.UpdateTime = updateTime

	return commission, nil
}
