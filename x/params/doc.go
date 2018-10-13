package params

/*
Package params provides a globally available parameter store.

There are two main types, Keeper and Subspace. Subspace is an isolated namespace for a
paramstore, where keys are prefixed by preconfigured spacename. Keeper has a
permission to access all existing spaces.

Subspace can be used by the individual keepers, who needs a private parameter store
that the other keeper cannot modify. Keeper can be used by the Governance keeper,
who need to modify any parameter in case of the proposal passes.

Basic Usage:

First, declare parameter space and parameter keys for the module. Then include
params.Subspace in the keeper. Since we prefix the keys with the spacename, it is
recommended to use the same name with the module's.

	const (
		DefaultParamspace = "mymodule"
	)

	const (
		KeyParameter1 = "myparameter1"
		KeyParameter2 = "myparameter2"
	)

	type Keeper struct {
		cdc *wire.Codec
		key sdk.StoreKey

		ps params.Subspace
	}

Pass a params.Subspace to NewKeeper with DefaultParamspace (or another)

	app.myKeeper = mymodule.NewKeeper(app.paramStore.SubStore(mymodule.DefaultParamspace))

Now we can access to the paramstore using Paramstore Keys

	var param MyStruct
	k.ps.Get(KeyParameter1, &param)
	k.ps.Set(KeyParameter2, param)

Genesis Usage:

Declare a struct for parameters and make it implement params.ParamSet. It will then
be able to be passed to SetParamSet.

	type MyParams struct {
		Parameter1 uint64
		Parameter2 string
	}

	// Implements params.ParamSet
	// KeyValuePairs must return the list of (ParamKey, PointerToTheField)
	func (p *MyParams) KeyValuePairs() params.KeyValuePairs {
		return params.KeyFieldPairs {
			{KeyParameter1, &p.Parameter1},
			{KeyParameter2, &p.Parameter2},
		}
	}

	func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
		k.ps.SetParamSet(ctx, &data.params)
	}

The method is pointer receiver because there could be a case that we read from
the store and set the result to the struct.

Master Keeper Usage:

Keepers that require master permission to the paramstore, such as gov, can take
params.Keeper itself to access all subspace(using GetSubspace)

	type MasterKeeper struct {
		pk params.Keeper
	}

	func (k MasterKeeper) SetParam(ctx sdk.Context, space string, key string, param interface{}) {
		space, ok := k.pk.GetSubspace(space)
		if !ok {
			return
		}
		space.Set(ctx, key, param)
	}
*/
