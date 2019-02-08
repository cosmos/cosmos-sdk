# Staking module specification

## Abstract

This paper specifies the Staking module of the Cosmos-SDK, which was first
described in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper)
in June 2016. 

The module enables Cosmos-SDK based blockchain to support an advanced
Proof-of-Stake system. In this system, holders of the native staking token of
the chain can become validators and can delegate tokens to validator
validators, ultimately determining the effective validator set for the system.

This module will be used in the Cosmos Hub, the first Hub in the Cosmos
network.

## Contents

The following specification uses *Atom* as the native staking token. The module
can be adapted to any Proof-Of-Stake blockchain by replacing *Atom* with the
native staking token of the chain.

 1. **[State](state.md)**	
   - Pool
   - Params
   - Validator
   - Delegation
   - UnbondingDelegation
   - Redelegation
   - Queues
 2. **[State Transistions](state_transitions.md)**	
   - Validator
   - Delegation
   - Slashing
 3. **[Messages](messages.md)**	
   - MsgCreateValidator
   - MsgEditValidator
   - MsgDelegate
   - MsgBeginUnbonding
   - MsgBeginRedelegate
 4. **[End-Block](end_block.md)**	
   - Validator Set Changes
   - Queues 
   - Unbonding Validators
   - Unbonding Delegations
   - Redelegations
 5. **[Hooks](hooks.md)**	
 6. **[Tags](tags.md)**	

