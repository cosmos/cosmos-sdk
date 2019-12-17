<!--
order: 1
-->

# Concepts

## Evidence

Any concrete type of evidence submitted to the `x/evidence` module must fulfill the
`Evidence` contract outlined below. Not all concrete types of evidence will fulfill
this contract in the same way and some data may be entirely irrelevant to certain
types of evidence.

```go
type Evidence interface {
  Route() string
  Type() string
  String() string
  Hash() HexBytes
  ValidateBasic() error

  // The consensus address of the malicious validator at time of infraction
  GetConsensusAddress() ConsAddress

  // Height at which the infraction occurred
  GetHeight() int64

  // The total power of the malicious validator at time of infraction
  GetValidatorPower() int64

  // The total validator set power at time of infraction
  GetTotalPower() int64
}
```

## Registration & Handling

The `x/evidence` module must first know about all types of evidence it is expected
to handle. This is accomplished by registering the `Route` method in the `Evidence`
contract with what is known as a `Router` (defined below). The `Router` accepts
`Evidence` and attempts to find the corresponding `Handler` for the `Evidence`
via the `Route` method.

```go
type Router interface {
  AddRoute(r string, h Handler) Router
  HasRoute(r string) bool
  GetRoute(path string) Handler
  Seal()
  Sealed() bool
}
```

The `Handler` (defined below) is responsible for executing the entirety of the
business logic for handling `Evidence`. This typically includes validating the
evidence, both stateless checks via `ValidateBasic` and stateful checks via any
keepers provided to the `Handler`. In addition, the `Handler` may also perform
capabilities such as slashing and jailing a validator.

```go
type Handler func(Context, Evidence) error
```
