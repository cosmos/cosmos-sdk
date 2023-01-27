package examples

import (
	"context"
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

var validatorCodec collections.ValueCodec[Validator] = nil // faux collections.ValueCodec, this is most likely a proto so you'll be using codec.ProtoValueCodec

type Validator struct {
	Power   uint64
	Moniker string
}

type ValidatorsIndexes struct {
	// Power is the index of validators by their power.
	Power *indexes.Multi[uint64, ValAddress, Validator]
}

// IndexesList implements collections.Indexes
func (i ValidatorsIndexes) IndexesList() []collections.Index[ValAddress, Validator] {
	return []collections.Index[ValAddress, Validator]{i.Power}
}

type Keeper2 struct {
	Validators *collections.IndexedMap[ValAddress, Validator, ValidatorsIndexes]
}

func NewKeeper2() Keeper2 {
	sb := collections.NewSchemaBuilderFromKVService(nil)
	return Keeper2{
		Validators: collections.NewIndexedMap(
			sb,
			collections.NewPrefix(0),
			"validators",
			collections.StringKey, // this is how we encode the key, remember ValAddress is an alias of string
			validatorCodec,
			ValidatorsIndexes{
				Power: indexes.NewMulti(
					sb, collections.NewPrefix(1), "validators_by_power",
					collections.Uint64Key, // we pass how we encode the Power field, which is uint64
					collections.StringKey, // we pass how we encode the primary key which is a string key, the same we used above.
					func(_ ValAddress, value Validator) (uint64, error) { // we pass a function that given the Validator struct returns their power
						return value.Power, nil
					},
				),
			},
		),
	}
}

// this returns all the validators with power equal to the one specified
func (k Keeper2) GetValidatorWithPowerEqualTo(ctx context.Context, power uint64) ([]Validator, error) {
	iter, err := k.Validators.Indexes.Power.MatchExact(ctx, power)
	if err != nil {
		return nil, err
	}
	return indexes.CollectValues(ctx, k.Validators, iter)
}
