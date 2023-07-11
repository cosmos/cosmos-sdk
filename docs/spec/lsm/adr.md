# ADR: Liquid Staking

## Changelog

* 2022-09-10: Initial Draft (@zmanian)
* 2023-07-10:  (@zmanian, @sampocs, @rileyedmunds, @mpoke)

## Abstract

Add a semi-fungible liquid staking primitive to the default Cosmos SDK staking module. While implemented as changes to existing modules, these additional features are hereinafter referred to as the liquid staking module (LSM). This upgrades proof of stake to enable safe designs with lower overall monetary issuance and integration with numerous liquid staking protocols like Stride, Persistence, Quicksilver, Lido etc.

## Context

The original release of the Cosmos Hub featured the implementation of a ground breaking proof of stake mechanism featuring delegation, slashing, in protocol reward distribution and adaptive issuance. This design was state of the art for 2016 and has been deployed without major changes by many L1 blockchains.

As both Proof of Stake and blockchain use cases have matured, this design has aged poorly and should no longer be considered a good baseline Proof of Stake issuance. In the world of application specific blockchains, there cannot be a one size fits all blockchain but the Cosmos SDK does endeavour to provide a good baseline implementation and one that is suitable for the Cosmos Hub.

The most important deficiency of the legacy staking design is that it composes poorly with on chain protocols for trading, lending, derivatives that are referred to collectively as DeFi. The legacy staking implementation starves these applications of liquidity by increasing the risk free rate adaptively. It basically makes DeFi and staking security somewhat incompatible. 

The Osmosis team has adopted the idea of Superfluid and Interfluid staking where assets that are participating in DeFi appliactions can also be used in proof of stake. This requires tight integration with an enshrined set of DeFi applications and thus is unsuitable for the Cosmos SDK.

It's also important to note that Interchain Accounts are available in the default IBC implementation and can be used to [rehypothecate](https://www.investopedia.com/terms/h/hypothecation.asp#toc-what-is-rehypothecation) delegations. Thus liquid staking is already possible and these changes merely improve the UX of liquid staking. Centralized exchanges also rehypothecate staked assets, posing challenges for decentralization. This ADR takes the position that adoption of in-protocol liquid staking is the preferable outcome and provides new levers to incentivize decentralization of stake. 

These changes to the staking module have been in development for more than a year and have seen substantial industry adoption by protocols who plan to build staking UX. The internal economics at Informal team has also done a review of the impacts of these changes and this review led to the development of the validator bond system. This system provides governance with a tuneable parameter for modulating the risks of principal agent problem called the validator bond factor. 

Liquid proof of stake systems exacerbate the risk that a single entity - the liquid staking provider - amasses more than ⅓ the total staked supply on a given chain, giving it the power to halt that chain’s block production or censor transactions and proposals.

Liquid proof of stake may also exacerbates the principal agent risk that exists at the heart of the delegated proof of stake system. The core of the problem is that validators do not actually own the stake that is delegated to them. This leaves the open to perverse incentives to attack the consensus system. Cosmos introduced the idea of min self bond in the staking. This creates a minimum amount of stake the must be bonded by the validators operator key. This feature has very little effect on the behavior of delegates.

## Decision

We implement the semi-fungible liquid staking system and validator bond factor system within the cosmos sdk. Though registered as fungible assets, these tokenized shares have extremely limited fungibility, only among the specific delegation record that was created when shares were tokenized. These assets can be used for OTC trades but composability with DeFi is limited. The primary expected use case is improving the user experience of liquid staking providers.

The LSM is designed to safely and efficiently facilitate the adoption of liquid staking.

The LSM mitigates liquid staking risks by limiting the total amount of tokens that can be liquid staked to 25% of all staked tokens. 

As additional risk-mitigation features, the LSM introduces a requirement that validators self-bond tokens to be eligible for delegations from liquid staking providers, and that the portion of their liquid staked shares must not exceed 50% of their total shares.

A new governance parameter is introduced that defines the ratio of validator bonded tokens to issued tokenized shares. This is called the validator bond factor. A larger validator bond factor allows more tokenized shares to be issued for a smaller amount of validator bond. If governance is comfortable with how the liquid staking market is evolving, it makes sense to increase this value.

Min self delegation is removed from the staking system with the expectation that it will be replaced by the validator bond system. The validator bond system allows multiple accounts to demonstrate economic alignment with the validator operator as team members, partners etc. without co-mingling funds. Validator bonding will likely be required to grow the validators' business under widespread adoption of liquid staking once governance has adjusted the validator bond factor.

When shares are tokenized, the underlying shares are transferred to a module account and rewards go to the module account for the TokenizedShareRecord. 

There is no longer a mechanism to override the validators vote for TokenizedShares.


### Limiting liquid staking


The LSM would limit the percentage of liquid staked tokens by all liquid staking providers to 25% of the total supply of staked tokens. For example, if 100M tokens were currently staked, and if the LSM were installed today then the total liquid staked supply would be limited to a maximum of 25M tokens.

This is a key safety feature, as it would prevent liquid staking providers from collectively controlling more than ⅓ of the total staked token supply, which is the threshold at which a group of bad actors could halt block production.

Additionally, a separate cap is enforced on each validator's portion of liquid staked shares. Once 50% of shares are liquid, the validator is unable to accept additional liquid stakes.

Technically speaking, this cap on liquid staked tokens is enforced by limiting the total number of tokens that can be staked via interchain accounts plus the number of tokens that can be tokenized using LSM. Once this joint cap is reached, the LSM prevents interchain accounts from staking any more tokens and prevents tokenization of delegations using LSM.


### Validator bond

As an additional security feature, validators who want to receive delegations from liquid staking providers would be required to self-bond a certain amount of tokens. The validator self-bond, or “validator-bond,” means that validators need to have “skin in the game” in order to be entrusted with delegations from liquid staking providers. This disincentivizes malicious behavior and enables the validator to negotiate its relationship with liquid staking providers.

Technically speaking, the validator-bond is tracked by the LSM. The maximum number of tokens that can be delegated to a validator by a liquid staking provider is equal to the validator-bond multiplied by the “validator-bond factor.” The initial validator bond factor would be set at 250, but can be configured by governance. 

With a validator-bond factor of 250, for every 1 token a validator self-bonds, that validator is eligible to receive up to two-hundred-and-fifty tokens delegated from liquid staking providers. The validator-bond has no impact on anything other than eligibility for delegations from liquid staking providers.

Without self-bonding tokens, a validator can’t receive delegations from liquid staking providers. And if a validator’s maximum amount of delegated tokens from liquid staking providers has been met, it would have to self-bond more tokens to become eligible for additional liquid staking provider delegations.

### Instantly liquid staking tokens that are already staked

Next, let’s discuss how the LSM makes the adoption of liquid staking more efficient, and can help the blockchain that installs it build strong relationships with liquid staking providers. The LSM enables users to instantly liquid stake their staked tokens, without having to wait the twenty-one day unbonding period. This is important, because a very large portion of the token supply on most Cosmos blockchains is currently staked. Liquid staking tokens that are already staked incur a switching cost in the form of forfeited staking rewards over the chain's unbonding period. The LSM eliminates this switching cost.


A user would be able to visit any liquid staking provider that has integrated with the LSM and click a button to convert his staked tokens to liquid staked tokens. It would be as easy as liquid staking unstaked tokens.

Technically speaking, this is accomplished by using something called an “LSM share.” Using the liquid staking module, a user can tokenize their staked tokens and turn it into LSM shares. LSM shares can be redeemed for underlying staked tokens and are transferable. After staked tokens are tokenized they can be immediately transferred to a liquid staking provider in exchange for liquid staking tokens - without having to wait for the unbonding period.

## Toggling the ability to tokenize shares

Currently LSM facilitates the immediate conversion of staked assets into liquid staked tokens (referred to as "tokenization"). Despite the many benefits that come with this capability, it does inadvertently negate a protective measure available via traditional staking, where a user can stake their tokens to render them illiquid in the event that their wallet is compromised (the attacker would first need to unbond, then transfer out the tokens).

LSM would obviate this safety measure, as an attacker could tokenize and immediately transfer staked tokens to another wallet. So, as an additional protective measure, this proposal incorporates a feature to permit users to selectively disable the tokenization of their stake. 

The LSM grants the ability to enable and disable the ability to tokenizate their stake. When tokenization is disabled, a lock is placed on the user's account, effectively preventing the conversion of any of their delegations. Re-enabling tokenization would initiate the removal of the lock, but the process is not immediate. The lock removal is queued, with the lock itself persisting throughout the unbonding period. Following the completion of the unbonding period, the lock would be completely removed, restoring the user's ablility to tokenize. For users who choose to enable the lock, this delay better positions them to regain control of their funds in the event their wallet is compromised.

## Economics

We expect that eventually governance may decide that the principal agent problems between validators and liquid staking are resolved through the existence of mature liquid staking synthetic asset systems and their associate risk framework. Governance can effectively disable the feature by setting the scalar value to -1 and allow unlimited minting and all liquid delegations to be freely undelegated.

During the transitionary period, this creates a market for liquid shares that may serve to help further decentralize the validator set.

It also allows multiple participants in a validator business to hold their personal stakes in segregated accounts but all collectively contribute towards demonstrating alignment with the safety of the protocol.

## Instructions for validators
Once delegated to a validator, a delegator (or validator operator) can convert their delegation to a validator into Validator Bond by signing a ValidatorBond message. 

The ValidatorBond message is exposed by the staking module and can be executed as follows:
```
gaiad tx staking validator-bond cosmosvaloper13h5xdxhsdaugwdrkusf8lkgu406h8t62jkqv3h <delegator> --from mykey  
```
There are no partial Validator Bonds: when a delegator or validator converts their shares to a particular validator into Validator Bond, their entire delegation to that validator is converted to Validator Bond. If a validator or delegator wishes to convert only some of their delegation to Validator Bond, they should transfer those funds to a separate address and Validator Bond from that address, or redelegate the funds that they do not wish to validator bond to another validator before converting their delegation to validator bond.

To convert Validator Bond back into a standard delegation, simply unbond the shares.

## Technical Spec:

 Please see this document for a technical spec for the LSM:
 https://docs.google.com/document/d/1WYPUHmQii4o-q2225D_XyqE6-1bvM7Q128Y9amqRwqY/edit#heading=h.zcpx47mn67kl

### Software parameters

New governance parameters are introduced that define the cap on the percentage of delegated shares than can be liquid, namely the `GlobalLiquidStakingCap` and `ValidatorLiquidStakingCap`. The `ValidatorBondFactor` governance parameter defines the number of tokens that can be liquid staked, relative to a validator's validator bond.

```proto
// Params defines the parameters for the staking module.
message Params {
  // ... existing params...
  // validator_bond_factor is required as a safety check for tokenizing shares and 
  // delegations from liquid staking providers
  string validator_bond_factor = 7 [
    (gogoproto.moretags) = "yaml:\"validator_bond_factor\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable) = false
  ];
  // global_liquid_staking_cap represents a cap on the portion of stake that 
  // comes from liquid staking providers
  string global_liquid_staking_cap = 8 [
    (gogoproto.moretags)   = "yaml:\"global_liquid_staking_cap\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
  // validator_liquid_staking_cap represents a cap on the portion of stake that 
  // comes from liquid staking providers for a specific validator
  string validator_liquid_staking_cap = 9 [
    (gogoproto.moretags)   = "yaml:\"validator_liquid_staking_cap\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
}
```

### Data structures

#### Validator
The `TotalValidatorBondShares` and `TotalLiquidShares` attributes were added to the `Validator` struct.

```proto
message Validator {
  // ...existing attributes...
  // Number of shares self bonded from the validator
  string total_validator_bond_shares = 11 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
  // Total number of shares either tokenized or owned by a liquid staking provider 
  string total_liquid_shares = 12 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
}
```

#### Delegation
The `ValidatorBond` attribute was added to the `Delegation` struct.

```proto
// Delegation represents the bond with tokens held by an account. It is
// owned by one delegator, and is associated with the voting power of one
// validator.
message Delegation {
  // ...existing attributes...
  // has this delegation been marked as a validator self bond.
  bool validator_bond = 4;
}
```

#### Toggling the ability to tokenize shares
```proto
// PendingTokenizeShareAuthorizations stores a list of addresses that have their 
// tokenize share re-enablement in progress
message PendingTokenizeShareAuthorizations {
  repeated string addresses = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
// Prevents an address from tokenizing any of their delegations
message MsgDisableTokenizeShares {
  string delegator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

// EnableTokenizeShares begins the re-allowing of tokenizing shares for an address,
// which will complete after the unbonding period
// The time at which the lock is completely removed is returned in the response
message MsgEnableTokenizeShares {
  string delegator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```


### Tracking total liquid stake
To monitor the progress towards the global liquid staking cap, the module needs to know two things: the total amount of staked tokens and the total amount of *liquid staked* tokens. The total staked tokens can be found by checking the balance of the "Bonded" pool. The total *liquid staked* tokens are stored separately and can be found under the `TotalLiquidStakedTokensKey` prefix (`[]byte{0x65}`). The value is managed by the following keeper functions:
```go
func (k Keeper) SetTotalLiquidStakedTokens(ctx sdk.Context, tokens sdk.Dec)
func (k Keeper) GetTotalLiquidStakedTokens(ctx sdk.Context) sdk.Dec
```

### Tokenizing shares

The MsgTokenizeShares message is used to create tokenize delegated tokens. This message can be executed by any delegator who has positive amount of delegation and after execution the specific amount of delegation disappear from the account and share tokens are provided. Share tokens are denominated in the validator and record id of the underlying delegation.

A user may tokenize some or all of their delegation.

They will receive shares with the denom of `cosmosvaloper1xxxx/5` where 5 is the record id for the validator operator.

MsgTokenizeShares fails if the account is a VestingAccount. Users will have to move vested tokens to a new account and endure the unbonding period. We view this as an acceptable tradeoff vs. the complex book keeping required to track vested tokens.

The total amount of outstanding tokenized shares for the validator is checked against the sum of validator bond delegations multiplied by the validator bond factor. If the tokenized shares exceeds this limit, execution fails.

MsgTokenizeSharesResponse provides the number of tokens generated and their denom.

### Helper functions
In order to identify whether a liquid stake transaction will exceed either the global liquid staking cap or the validator bond cap, the following functions were added:

```go
// Check if an account is a owned by a liquid staking provider
// This is determined by checking if the account is a 32-length module account
func (k Keeper) DelegatorIsLiquidStaker(address sdk.AccAddress) bool 

// SafelyIncreaseTotalLiquidStakedTokens increments the total liquid staked tokens
// if the caps are enabled and the global cap is not surpassed by this delegation
func (k Keeper) SafelyIncreaseTotalLiquidStakedTokens(ctx sdk.Context, amount sdk.Int) error 

// DecreaseTotalLiquidStakedTokens decrements the total liquid staked tokens
// if the caps are enabled
func (k Keeper) DecreaseTotalLiquidStakedTokens(ctx sdk.Context, amount sdk.Int) error

// SafelyIncreaseValidatorTotalLiquidShares increments the total liquid shares on a validator
// if the caps are enabled and the validator bond cap is not surpassed by this delegation
func (k Keeper) SafelyIncreaseValidatorTotalLiquidShares(ctx sdk.Context, validator types.Validator, shares sdk.Dec) error 

// DecreaseValidatorTotalLiquidShares decrements the total liquid shares on a validator
// if the caps are enabled
func (k Keeper) DecreaseValidatorTotalLiquidShares(ctx sdk.Context, validator types.Validator, shares sdk.Dec) error

// SafelyDecreaseValidatorBond decrements the total validator's self bond
// so long as it will not cause the current delegations to exceed the threshold
// set by validator bond factor
func (k Keeper) SafelyDecreaseValidatorBond(ctx sdk.Context, validator types.Validator, shares sdk.Dec) error 
```

### Accounting
Tracking the total liquid stake and total liquid validator shares requires additional accounting changes in the following transactions/events:

```go
func Delegate() {
    ...
    // If delegator is a liquid staking provider
    //    Increment total liquid staked
    //    Increment validator liquid shares
}

func Undelegate() {
    ...
    // If delegator is a liquid staking provider
    //    Decrement total liquid staked
    //    Decrement validator liquid shares
}

func BeginRedelegate() {
    ...
    // If delegator is a liquid staking provider
    //    Decrement source validator liquid shares
    //    Increment destination validator liquid shares
}

func TokenizeShares() {
    ...
    // If delegator is a NOT liquid staking provider (otherwise the shares are already included)
    //    Increment total liquid staked
    //    Increment validator liquid shares
}

func RedeemTokens() {
    ...
    // If delegator is a NOT liquid staking provider 
    //    Decrement total liquid staked
    //    Decrement validator liquid shares
}

func Slash() {
    ...
    // Decrement total liquid staked (since total liquid stake is denominated in tokens, not shares)
}
```

### Transaction failure cases
With the liquid staking caps in consideration, there are additional scenarios that should cause a transaction to fail:
```go

func Delegate() {
    ...
    // If delegator is a liquid staking provider
    //    Fail transaction if delegation exceeds global liquid staking cap
    //    Fail transaction if delegation exceeds validator liquid staking cap
    //    Fail transaction if delegation exceeds validator bond cap
}

func Undelegate() {
    ...
    // If the unbonded delegation is a ValidatorBond
    //    Fail transaction if the reduction in validator bond would cause the
    //    existing liquid delegation to exceed the cap
}

func BeginRedelegate() {
    ...
    // If the delegation is a ValidatorBond
    //    Fail transaction if the reduction in validator bond would cause the
    //    existing liquid delegation to exceed the cap

    // If delegator is a liquid staking provider
    //    Fail transaction if delegation exceeds global liquid staking cap
    //    Fail transaction if delegation exceeds validator liquid staking cap
    //    Fail transaction if delegation exceeds validator bond cap
}

func TokenizeShares() {
    ...
    // If the delegation is a ValidatorBond
    //    Fail transaction - ValidatorBond's cannot be tokenized

    // If the sender is NOT a liquid staking provider
    //    Fail transaction if tokenized shares would exceed the global liquid staking cap
    //    Fail transaction if tokenized shares would exceed the validator liquid staking cap
    //    Fail transaction if tokenized shares would exceed the validator bond cap
}
```

### Bootstrapping total liquid stake
When upgrading to enable the liquid staking module, the total global liquid stake and total liquid validator shares must be determined. This can be done in the upgrade handler by looping through delegation records and including the delegation in the total if the delegator has a 32-length address. This is implemented by the following function:
```go
func RefreshTotalLiquidStaked() {
  // Resets all validator TotalLiquidShares to 0
  // Loops delegation records
  //    For each delegation, determines if the delegation was from a 32-length address
  //    If so, increments the global liquid staking cap and validator liquid shares
}
```

### Toggling the ability to tokenize shares

```go
// Adds a lock that prevents tokenizing shares for an account
// The tokenize share lock store is implemented by keying on the account address
// and storing a timestamp as the value. The timestamp is empty when the lock is
// set and gets populated with the unlock completion time once the unlock has started
func AddTokenizeSharesLock(address sdk.AccAddress) 

// Removes the tokenize share lock for an account to enable tokenizing shares
func RemoveTokenizeSharesLock(address sdk.AccAddress) 

// Updates the timestamp associated with a lock to the time at which the lock expires
func SetTokenizeShareUnlockTime(address sdk.AccAddress, completionTime time.Time) 

// Checks if there is currently a tokenize share lock for a given account
// Returns a bool indicating if the account is locked, as well as the unlock time
// which may be empty if an unlock has not been initiated
func IsTokenizeSharesDisabled(address sdk.AccAddress) (disabled bool, unlockTime time.Time) 

// Stores a list of addresses pending tokenize share unlocking at the same time
func SetPendingTokenizeShareAuthorizations(completionTime time.Time, authorizations types.PendingTokenizeShareAuthorizations)

// Returns a list of addresses pending tokenize share unlocking at the same time
func GetPendingTokenizeShareAuthorizations() PendingTokenizeShareAuthorizations 

// Inserts the address into a queue where it will sit for 1 unbonding period
// before the tokenize share lock is removed
// Returns the completion time
func QueueTokenizeSharesAuthorization(address sdk.AccAddress) time.Time 

// Unlocks all queued tokenize share authorizations that have matured
// (i.e. have waited the full unbonding period)
func RemoveExpiredTokenizeShareLocks(blockTime time.Time) (unlockedAddresses []string) 
```