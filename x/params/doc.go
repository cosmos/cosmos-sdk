package params

/*
Package params provides a globally available parameter store.

There are two main types, Keeper and Space. Space is an isolated namespace for a
paramstore, where keys are prefixed by preconfigured spacename. Keeper has a
permission to access all existing spaces and create new space.

Space can be used by the individual keepers, who needs a private parameter store
that the other keeper are not able to modify. Keeper can be used by the Governance
keeper, who need to modify any parameter in case of the proposal passes.

Basic Usage:

First, declare parameter space and parameter keys for the module. Then include
params.Store in the keeper. Since we prefix the keys with the spacename, it is
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

		ps params.Store
	}

Pass a params.Store to NewKeeper with DefaultParamSpace (or another)

	app.myKeeper = mymodule.NewKeeper(app.paramStore.SubStore(mymodule.DefaultParamspace))

Now we can access to the paramstore using Paramstore Keys

	k.ps.Get(KeyParameter1, &param)
	k.ps.Set(KeyParameter2, param)

Genesis Usage:

Declare a struct for parameters and make it implement ParamStruct. It will then
be able to be passed to SetFromParamStruct.

	type MyParams struct {
		Parameter1 uint64
		Parameter2 string
	}

	func (p *MyParams) KeyFieldPairs() params.KeyFieldPairs {
		return params.KeyFieldPairs {
			{KeyParameter1, &p.Parameter1},
			{KeyParameter2, &p.Parameter2},
		}
	}

	func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
		k.ps.SetFromParamStruct(ctx, &data.params)
	}

The method is pointer receiver because there could be a case that we read from
the store and set the result to the struct.

Master Permission Usage:

Keepers that require master permission to the paramstore, such as gov, can take
params.Keeper itself to access all substores(using GetSubstore)

	type MasterKeeper struct {
		ps params.Store
	}

	func (k MasterKeeper) SetParam(ctx sdk.Context, space string, key string, param interface{}) {
		store, ok := k.ps.GetSubstore(space)
		if !ok {
			return
		}
		store.Set(ctx, key, param)
	}
*/
