## Application codec

**File: [`app/app.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go)**

Finally, we need to define the `MakeCodec()` function and register the concrete types and interface from the various modules.

```go
func MakeCodec() *wire.Codec {
    var cdc = wire.NewCodec()
    wire.RegisterCrypto(cdc) // Register crypto.
    sdk.RegisterWire(cdc)    // Register Msgs
    bank.RegisterWire(cdc)
    simplestake.RegisterWire(cdc)
    simpleGov.RegisterWire(cdc)

    // Register AppAccount
    cdc.RegisterInterface((*auth.Account)(nil), nil)
    cdc.RegisterConcrete(&types.AppAccount{}, "simpleGov/Account", nil)
    return cdc
}
```