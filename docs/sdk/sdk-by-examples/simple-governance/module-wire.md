## Wire 

**File: [`x/simple_governance/wire.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/wire.go)**

The `wire.go` file allows developers to register the concrete message types of their module into the codec. In our case, we have two messages to declare:

```go
func RegisterWire(cdc *wire.Codec) {
    cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
    cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)
}
```
Don't forget to call this function in `app.go` (see [Application - Bridging it all together](app-structure.md)) for more).