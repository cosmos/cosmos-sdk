package simulation

import (
	"bytes"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/types"

	"cosmossdk.io/core/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding staking type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.LastTotalPowerKey):
			var powerA, powerB math.Int
			if err := powerA.Unmarshal(kvA.Value); err != nil {
				panic(err)
			}
			if err := powerB.Unmarshal(kvB.Value); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", powerA, powerB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorsKey):
			var validatorA, validatorB types.Validator

			if err := cdc.Unmarshal(kvA.Value, &validatorA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &validatorB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", validatorA, validatorB)
		case bytes.Equal(kvA.Key[:1], types.LastValidatorPowerKey),
			bytes.Equal(kvA.Key[:1], types.ValidatorsByConsAddrKey),
			bytes.Equal(kvA.Key[:1], types.ValidatorsByPowerIndexKey):
			return fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.DelegationKey):
			var delegationA, delegationB types.Delegation

			if err := cdc.Unmarshal(kvA.Value, &delegationA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &delegationB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", delegationA, delegationB)
		case bytes.Equal(kvA.Key[:1], types.UnbondingDelegationKey),
			bytes.Equal(kvA.Key[:1], types.UnbondingDelegationByValIndexKey):
			var ubdA, ubdB types.UnbondingDelegation

			if err := cdc.Unmarshal(kvA.Value, &ubdA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &ubdB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", ubdA, ubdB)
		case bytes.Equal(kvA.Key[:1], types.RedelegationKey),
			bytes.Equal(kvA.Key[:1], types.RedelegationByValSrcIndexKey):
			var redA, redB types.Redelegation

			if err := cdc.Unmarshal(kvA.Value, &redA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &redB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", redA, redB)
		case bytes.Equal(kvA.Key[:1], types.ParamsKey):
			var paramsA, paramsB types.Params

			if err := cdc.Unmarshal(kvA.Value, &paramsA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &paramsB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", paramsA, paramsB)
		case bytes.Equal(kvA.Key[:1], types.ValidatorConsPubKeyRotationHistoryKey):
			var historyA, historyB types.ConsPubKeyRotationHistory

			if err := cdc.Unmarshal(kvA.Value, &historyA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &historyB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", historyA, historyB)
		case bytes.Equal(kvA.Key[:1], types.ValidatorConsensusKeyRotationRecordQueueKey):
			var historyA, historyB types.ValAddrsOfRotatedConsKeys

			if err := cdc.Unmarshal(kvA.Value, &historyA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &historyB); err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", historyA, historyB)
		default:
			panic(fmt.Sprintf("invalid staking key prefix %X", kvA.Key[:1]))
		}
	}
}
