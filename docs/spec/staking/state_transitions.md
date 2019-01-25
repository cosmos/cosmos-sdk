# State Transitions

This document describes the state transition operations pertaining to:
 - Validators
 - Delegations
 - Slashing

## Validators
RemoveValidatorTokens
RemoveValidatorTokensAndShares
AddValidatorTokensAndShares

completeUnbondingValidator
beginUnbondingValidator
bondValidator

bondedToUnbonding
 -> beginUnbondingValidator
unbondingToBonded
 -> bondValidator
unbondedToBonded
 -> bondValidator
unbondingToUnbonded
 -> completeUnbondingValidator

jailValidator / unjailValidator

## Delegations
Delegate
unbond
Undelegate
CompleteUnbonding
BeginRedelegation
CompleteRedelegation

## Slashing
Slash
slashUnbondingDelegation
slashRedelegation

