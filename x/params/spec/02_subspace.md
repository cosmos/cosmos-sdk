<!--
order: 2
-->

# Subspace

`Subspace` is a prefixed subspace of the parameter store. Each module who use the 
parameter store will take a `Subspace`, not the `Keeper`, to isolate permission to access. 

## Key

Parameter keys are human readable alphanumeric strings. A parameter for the key 
`"ExampleParameter"` is stored under `[]byte("SubspaceName" + "/" + "ExampleParameter")`, 
	where `"SubspaceName"` is the name of the subspace.

Subkeys are secondary parameter keys those are used along with a primary parameter key.
Subkeys can be used for grouping or dynamic parameter key generation during runtime.

## KeyTable

All of the parameter keys that will be used should be registered at the compile 
time. `KeyTable` is essentially a `map[string]attribute`, where the `string` is a parameter key.

Currently, `attribute` only consists of `reflect.Type`, which indicates the parameter 
type. It is needed even if the state machine has no error, because the paraeter 
can be modified externally, for example via the governance.

Only primary keys have to be registered on the `KeyTable`. Subkeys inherit the 
attribute of the primary key.

## ParamSet

Modules often define a struct of parameters. Instead of calling methods with 
each of those parameters, when the struct implements `ParamSet`, it can be used 
with the following methods:

* `KeyTable.RegisterParamSet()`: registers all parameters in the struct
* `Subspace.{Get, Set}ParamSet()`: Get to & Set from the struct

The implementor should be a pointer in order to use `GetParamSet()`
