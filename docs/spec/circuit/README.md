# Circuit Breaker

Circuit breaker disables certain messages types/names when the corresponding parameter are set `true`. The parameter can be set by governance or automatic panic handler, for example. 

## Parameter 

Two keys are initially defined:

```go
var (
    MsgTypeKey = []byte("msgtype")
    MsgNameKey = []byte("msgname")
)
```

Both are set with `bool` type in `ParamKeyTable()`. The actual types/names are provided as subkey when accessing to the paramstore.

## Antehandler

The antehandler checks the type and the name of all msgs that is processes. If any of the parameter is set `true`, the antehandler aborts and the transaction fails.

## Genesis

`GenesisState` defines initial circuit breaked msg types/names. They are `nil` by default(no msg is circuit breaked).
