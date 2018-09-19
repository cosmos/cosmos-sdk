## Application codec

**File: [`app/app.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go)**

Finally, we need to define the `MakeCodec()` function and register the concrete types and interface from the various modules.

```go
func MakeCodec() *codec.Codec {
    var cdc = codec.New()
    codec.RegisterCrypto(cdc) // Register crypto.
    sdk.RegisterCodec(cdc)    // Register Msgs
    bank.RegisterCodec(cdc)
    simplestake.RegisterCodec(cdc)
    simpleGov.RegisterCodec(cdc)

    // Register AppAccount
    cdc.RegisterInterface((*auth.Account)(nil), nil)
    cdc.RegisterConcrete(&types.AppAccount{}, "simpleGov/Account", nil)
    return cdc
}
```