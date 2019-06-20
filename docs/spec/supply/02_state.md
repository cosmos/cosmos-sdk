# State

## Supply

The `Supply` is a passive tracker of the total supply of the chain:

- Supply: `0x0 -> amino(Supply)`

```go
type Supply struct {
	Total sdk.Coins // total supply of tokens registered on the chain
}
```
