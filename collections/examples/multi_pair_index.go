package examples

import (
	"context"
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

// we show an example of a multi pair index.
// which is an index that can be used for creating
// relationships between the second part of the pair key.

type ValAddress = string
type AccAddress = string

var delegationCodec collections.ValueCodec[Delegation] = nil

type Delegation = struct {
	Denom  string
	Amount uint64
}

type DelegationsIndexes struct {
	// Delegator is an index that allows us to find all the delegations of a specific delegator's AccAddress.
	Delegator *indexes.MultiPair[ValAddress, AccAddress, Delegation]
}

// IndexesList implements collections.Indexes
func (i DelegationsIndexes) IndexesList() []collections.Index[collections.Pair[ValAddress, AccAddress], Delegation] {
	return []collections.Index[collections.Pair[ValAddress, AccAddress], Delegation]{i.Delegator}
}

type Keeper1 struct {
	// Delegations are tracked with a key composed of the validator we are delegating to and the delegator,
	// it maps delegations. This is a collections.Pair key. Having it as a pair allows us to get all the
	// delegations of a specific validator.
	Delegations *collections.IndexedMap[collections.Pair[ValAddress, AccAddress], Delegation, DelegationsIndexes]
}

func NewKeeper1() Keeper1 {
	sb := collections.NewSchemaBuilder(nil)
	// we create the primary key codec, note it uses string key because we pretend
	// AccAddress and ValAddress are strings and not bytes.
	primaryKeyCodec := collections.PairKeyCodec(collections.StringKey, collections.StringKey)
	return Keeper1{
		Delegations: collections.NewIndexedMap(
			sb,
			collections.NewPrefix(0),
			"delegations",
			primaryKeyCodec,
			delegationCodec,
			DelegationsIndexes{
				Delegator: indexes.NewMultiPair[Delegation]( // we type hint golang that the value of the indexed map is Delegation
					sb,                       // we pass the schema builder
					collections.NewPrefix(1), // the prefix
					"delegator_index",        // a human name
					primaryKeyCodec,          // and the primary key codec, which is a pair codec! now IndexedMap will index the objects by the second part of the key too!
				),
			},
		),
	}
}

func (k Keeper1) GetDelegationsOfValidator(ctx context.Context, val ValAddress) ([]Delegation, error) {
	iter, err := k.Delegations.Iterate(ctx, collections.NewPrefixedPairRange[ValAddress, AccAddress](val))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	delegations, err := iter.Values()
	if err != nil {
		return nil, err
	}
	return delegations, nil
}

func (k Keeper1) GetDelegationsOfDelegator(ctx context.Context, del AccAddress) ([]Delegation, error) {
	iter, err := k.Delegations.Indexes.Delegator.MatchExact(ctx, del)
	if err != nil {
		return nil, err
	}

	return k.Delegations.CollectValues(ctx, iter)
}
