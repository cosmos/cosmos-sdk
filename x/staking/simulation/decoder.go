package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding staking type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.LastTotalPowerKey):
			var powerA, powerB sdk.IntProto

			cdc.MustUnmarshalBinaryBare(kvA.Value, &powerA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &powerB)

			return fmt.Sprintf("%v\n%v", powerA, powerB)
		case bytes.Equal(kvA.Key[:1], types.ValidatorsKey):
			var validatorA, validatorB types.Validator

			cdc.MustUnmarshalBinaryBare(kvA.Value, &validatorA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &validatorB)

			return fmt.Sprintf("%v\n%v", validatorA, validatorB)
		case bytes.Equal(kvA.Key[:1], types.LastValidatorPowerKey),
			bytes.Equal(kvA.Key[:1], types.ValidatorsByConsAddrKey),
			bytes.Equal(kvA.Key[:1], types.ValidatorsByPowerIndexKey):
			return fmt.Sprintf("%v\n%v", sdk.ValAddress(kvA.Value), sdk.ValAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.DelegationKey):
			var delegationA, delegationB types.Delegation

			cdc.MustUnmarshalBinaryBare(kvA.Value, &delegationA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &delegationB)

			return fmt.Sprintf("%v\n%v", delegationA, delegationB)
		case bytes.Equal(kvA.Key[:1], types.UnbondingDelegationKey),
			bytes.Equal(kvA.Key[:1], types.UnbondingDelegationByValIndexKey):
			var ubdA, ubdB types.UnbondingDelegation

			cdc.MustUnmarshalBinaryBare(kvA.Value, &ubdA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &ubdB)

			return fmt.Sprintf("%v\n%v", ubdA, ubdB)
		case bytes.Equal(kvA.Key[:1], types.RedelegationKey),
			bytes.Equal(kvA.Key[:1], types.RedelegationByValSrcIndexKey):
			var redA, redB types.Redelegation

			cdc.MustUnmarshalBinaryBare(kvA.Value, &redA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &redB)

			return fmt.Sprintf("%v\n%v", redA, redB)
		default:
			panic(fmt.Sprintf("invalid staking key prefix %X", kvA.Key[:1]))
		}
	}
}
