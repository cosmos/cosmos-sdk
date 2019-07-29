# ADR 3: Module Sub-Accounts

## Changelog

## Context

Currently `ModuleAccount`s must be declared upon Supply Keeper initialization. In addition to this they don't allow for separation of fungible coins within an account.

We want to support the ability to define and manage sub-module-accounts.

## Decision

We will modify the existing `ModuleAccount` interface to support a heirarchical account structure.
The `ModuleAccount`s defined upon initialization of Supply Keeper are the roots of family trees.
Each `ModuleAccount` in a family tree may have zero or more children.
A `ModuleAccount` with one or more children is considered a parent to each child.
All `ModuleAccount`s have exactly one parent, unless they are the root of their family tree.

`ModuleAccount` permissions will be renamed to `Attribute`.
Each child's attributes must be a subset of their parent's attributes.
There is no limit on the number of children a `ModuleAccount` can have.
No `ModuleAccount`s can be removed from a family tree.
A `ModuleAccount` name must not contain ":".
A `ModuleAccount` address is the hash of its full path.
A `ModuleAccount`s full path is the path of the `ModuleAccount` names used to reach the child.
It starts with the root `ModuleAccount` name and is separated by a colon for each parent that follows until the child is reached.

Example name: `root:parent:child`

We will add a `TrackBalance` function which recursively updates the passive tracking of parent balances.
A `ModuleAccount` has no pubkeys.
The function `AddChildToModuleAccount` will be added to Supply Keeper, 
It will validate that the granted attributes are a subset of the parent and then register the child's name with the Supply Keeper.

### Implementation Changes

Modify `ModuleAccount` interface in `x/supply`:

```go
type ModuleAccount interface {
    GetName() string
    GetAddress() string
    GetAttribute() []string 
    HasAttribute() bool
    GetParent() string
    GetChildren() []string
    HasChild(string) bool 
    GetChildCoins() sdk.Coins
}
```

```go
// Implements the Account interface.
// ModuleAccount defines an account for modules that holds coins on a pool. A ModuleAccount
// may have sub-accounts known as children.
type ModuleAccount struct {
    *authtypes.BaseAccount
    Name        string    `json:"name" yaml:"name"`               // name of the module without the full path
    Attributes  []string  `json:"attributes" yaml:"attributes"`   // permissions of module account
    ChildCoins  sdk.Coins `json:"child_coins" yaml:"child_coins"` // passive tracking of sum of child balances
    Children    []string  `json:"children" yaml:"children"`       // array of children names
    Parent      string    `json:"parent" yaml:"parent"`           // full path of the parent name
}
```

```go
// Pseudocode
func TrackBalance(name string, delta sdk.Coins) {
    if name == "" {
        return
    } else {
        self.Balance += delta
    }
    TrackBalance(chopString(name)) // chop off last name after last ":"
    return
}
```

**Attributes**:

Attributes for a root `ModuleAccount` are declared upon initialization of the Supply Keeper.
A child `ModuleAccount` must have a subset of its parents attributes.

**Other changes**

We will add an invariant check for the `ModuleAccount` `GetChildCoins()` function, which will iterate over all `ModuleAccounts` to see if the sum of their child balances equals the passive tracking which is returned in `GetChildCoins()`

## Status

Accepted

## Consequences

### Positive

* `ModuleAccount` can separate fungible coins.
* `ModuleAccount` can dynamically add accounts.
* Children can have a subset of its parent's attributes.

### Negative

* `ModuleAccount` cannot be removed from a family tree.
* `ModuleAccount` implementation has increased complexity.

### Neutral

* `ModuleAccount` passively tracks child balances.

## References

Specs: [ModuleAccount](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/supply/01_concepts.md#module-accounts)

Issues: [4657](https://github.com/cosmos/cosmos-sdk/issues/4657)
