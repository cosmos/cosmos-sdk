package sims

import (
	"time"

	"github.com/cosmos/gogoproto/proto"
	any "github.com/cosmos/gogoproto/types/any"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StakingNotBondedPoolName = "not_bonded_tokens_pool"
	StakingBondedPoolName    = "bonded_tokens_pool"
)

type StakingGenesisState struct {
	// params defines all the parameters of related to deposit.
	Params StakingParams `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// last_total_power tracks the total amounts of bonded tokens recorded during
	// the previous end block.
	LastTotalPower math.Int `protobuf:"bytes,2,opt,name=last_total_power,json=lastTotalPower,proto3,customtype=cosmossdk.io/math.Int" json:"last_total_power"`
	// validators defines the validator set at genesis.
	Validators []StakingValidator `protobuf:"bytes,4,rep,name=validators,proto3" json:"validators"`
	// delegations defines the delegations active at genesis.
	Delegations []StakingDelegation `protobuf:"bytes,5,rep,name=delegations,proto3" json:"delegations"`
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
}

type StakingValidator struct {
	// operator_address defines the address of the validator's operator; bech encoded in JSON.
	OperatorAddress string `protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	// consensus_pubkey is the consensus public key of the validator, as a Protobuf Any.
	ConsensusPubkey *any.Any `protobuf:"bytes,2,opt,name=consensus_pubkey,json=consensusPubkey,proto3" json:"consensus_pubkey,omitempty"`
	// jailed defined whether the validator has been jailed from bonded status or not.
	Jailed bool `protobuf:"varint,3,opt,name=jailed,proto3" json:"jailed,omitempty"`
	// status is the validator status (bonded/unbonding/unbonded).
	Status int32 `protobuf:"varint,4,opt,name=status,proto3" json:"status,omitempty"`
	// tokens define the delegated tokens (incl. self-delegation).
	Tokens math.Int `protobuf:"bytes,5,opt,name=tokens,proto3,customtype=cosmossdk.io/math.Int" json:"tokens"`
	// delegator_shares defines total shares issued to a validator's delegators.
	DelegatorShares math.LegacyDec `protobuf:"bytes,6,opt,name=delegator_shares,json=delegatorShares,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"delegator_shares"`
	// min_self_delegation is the validator's self declared minimum self delegation.
	MinSelfDelegation math.Int `protobuf:"bytes,11,opt,name=min_self_delegation,json=minSelfDelegation,proto3,customtype=cosmossdk.io/math.Int" json:"min_self_delegation"`
}

type StakingDelegation struct {
	// delegator_address is the encoded address of the delegator.
	DelegatorAddress string `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	// validator_address is the encoded address of the validator.
	ValidatorAddress string `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	// shares define the delegation shares received.
	Shares math.LegacyDec `protobuf:"bytes,3,opt,name=shares,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"shares"`
}

func FakeStakingMsgCreateValidator(
	valAddr string, pubKey cryptotypes.PubKey, selfDelegation sdk.Coin,
	rate, maxRate, maxChangeRate math.LegacyDec, minSelfDelegation math.Int,
) (proto.Message, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}

	v := &StakingMsgCreateValidator{
		Commission: StakingValidatorCommission{
			Rate:          rate,
			MaxRate:       maxRate,
			MaxChangeRate: maxChangeRate,
		},
		MinSelfDelegation: minSelfDelegation,
		DelegatorAddress:  "",
		ValidatorAddress:  valAddr,
		Pubkey:            pkAny,
		Value:             selfDelegation,
	}

	_ = v

	return nil, nil
}

type StakingMsgCreateValidator struct {
	Commission        StakingValidatorCommission `protobuf:"bytes,2,opt,name=commission,proto3" json:"commission"`
	MinSelfDelegation math.Int                   `protobuf:"bytes,3,opt,name=min_self_delegation,json=minSelfDelegation,proto3,customtype=cosmossdk.io/math.Int" json:"min_self_delegation"`
	// Deprecated: Use of Delegator Address in MsgCreateValidator is deprecated.
	// The validator address bytes and delegator address bytes refer to the same account while creating validator (defer
	// only in bech32 notation).
	DelegatorAddress string   `protobuf:"bytes,4,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"` // Deprecated: Do not use.
	ValidatorAddress string   `protobuf:"bytes,5,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Pubkey           *any.Any `protobuf:"bytes,6,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
	Value            sdk.Coin `protobuf:"bytes,7,opt,name=value,proto3" json:"value"`
}

type StakingValidatorCommission struct {
	// rate is the commission rate charged to delegators, as a fraction.
	Rate math.LegacyDec `protobuf:"bytes,1,opt,name=rate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"rate"`
	// max_rate defines the maximum commission rate which validator can ever charge, as a fraction.
	MaxRate math.LegacyDec `protobuf:"bytes,2,opt,name=max_rate,json=maxRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"max_rate"`
	// max_change_rate defines the maximum daily increase of the validator commission, as a fraction.
	MaxChangeRate math.LegacyDec `protobuf:"bytes,3,opt,name=max_change_rate,json=maxChangeRate,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"max_change_rate"`
}
