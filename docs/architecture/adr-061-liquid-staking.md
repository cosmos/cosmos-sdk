# ADR ADR-061: Liquid Staking

## Changelog

* 2022-09-10: Initial Draft (@zmanian)

## Status

PROPOSED

## Abstract

Add a semi-fungible liquid staking primitive to the default cosmos SDK staking module. This upgrades proof of stake to enable safe designs with lower overall monetary issuance and integration with numerous liquid staking protocols like Stride, Persistence, Quicksilver, Lido etc.

## Context

The original release of the Cosmos Hub featured the implementation of a ground breaking proof of stake mechanism featuring delegation, slashing, in protocol reward distribution and adaptive issuance. This design was state of the art for 2016 and has been deployed without major changes by many L1 blockchains.

As both Proof of Stake and blockchain use cases have matured, this design has aged poorly and should no longer be considered a good baseline Proof of Stake issuance. In the world of application specific blockchains, there cannot be a one size fits all blockchain but the Cosmos SDK does endeavour to provide a good baseline implementation and one that is suitable for the Cosmos Hub.

The most important deficency of the legacy staking design is that it composes poorly with on chain protocols for trading, lending, derivatives that are referred to collectively as DeFi. The legacy staking implementation starves these applications of liquidity by increasing the risk free rate adaptively. It basically makes DeFi and staking security somewhat incompatible. 

The Osmosis team has adopted the idea of Superfluid and Interfluid staking where assets that are participating in DeFi appliactions can also be used in proof of stake. This requires tight integration with an enshrined set of DeFi applications and thus is unsuitable for the Cosmos SDK.

It's also important to note that Interchain Accounts are available in the default IBC implementation and can be used for staking and thus reyhypothneciation and liquid staking of staked assets is already possible. Thus liquid staking is already possible and these changes improve the UX of liquid staking. Centralized exahnges have also provided rehypthentication of staked assets and posed a challenge for decentralization. This ADR also take the position that liquid staking adoption is good and provides new levers to incentivize decentralization of stake. The

These changes to the staking module have been in development for more than a year and have seen substantial industry adoption who plan to build staking UX. The internal economics at Informal team has also done a review of the impacts of these changes and this review led to the developement of the exempt delegation system. This system provides governance with a tunenable parameter for modulating the risks of pricipal agent problem called the exemption factor. 

## Decision

We implement the semi-fungible liquid staking system and exemption factor system within the cosmos sdk.

A new governance parameter is introduced that defines the ratio of exempt to issued tokenized shares. This is called the exemption factor.

Min self delegation is removed from the staking system with the expectation that it will be replaced by the exempt delegations sytem.

When shares are tokenized, the underlying shares are transfered to a module account and rewards go to the module account for the TokenizedShareRecord. 

There is no longer a mechanism to override the validators vote for TokenizedShares.


### MsgTokenizeShares

The MsgTokenizeShares message is used to create tokenize delegated tokens. This message can be executed by any delegator who has positive amount of delegation and after execution the specific amount of delegation disappear from the account and share tokens are provided. Share tokens are demoninated in the validator and record id of the underlying delegation.

A user may tokenize some or all of their Delegation.

They will recieved shares like cosmosvaloper1xxxx5 where 5 is the record id for the validator operator.

A validator may tokenize their self bond but tokenizing more than their min self bond will be equivalent to unbonding their min self bond and cause the validator to be removed from the active set.

MsgTokenizeSharesResponse provides the number of tokens generated and their denom.


### MsgRedeemTokensforShares

The MsgRedeemTokensforShares message is used to redeem the delegation from share tokens. This message can be executed by any user who owns share tokens and after execution the delegation appear for the user.

### MsgTransferTokenizeShareRecord

The MsgTransferTokenizeShareRecord message is used to transfer the ownership of rewards generated from the tokenized amount of delegation. The tokenize share record is created when a user tokenize his/her delegation and deleted and full amount of share tokens are redeemed.

### MsgExemptDelegation

The MsgExemptDelegation message is used to exempt a delegation to a validator. If the exemption factor is greater than 0, this will enable more delegation to the validator


## Consequences


### Backwards Compatibility

By setting the exemption factor to zero, this module works like legacy staking. The only substantial change is the removal of min-self-bond and without any tokenized shares, there is no incentive to exempt delegation. 

### Positive

This approach should enable integration with liquid staking providers and improved user experience. It provides a pathway to security non-expoential issuance policites in the baseline staking module.


## References

