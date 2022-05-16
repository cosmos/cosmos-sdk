# ADR Proposer slashing for resource misuse attacks.

## Changelog


## Abstract

Block Proposers can construct blocks at their discretion but they are free to fill blocks with invalid transactions or exceed the block gas limit. This


## Context

During the Terra blockchain failure incident in May of 2022, we observed a very large number of invalid and duplicated transactions landing on chain. This provided the first concrete evidence that on a large Tendermint chain validators were running non-standard software to propose blocks. Protocols built with the cosmos sdk need to have slashing behavior that punish validators for poorly designed block construction software.


## Decision


Rational/Malicious proposers may commit protocol violations when using non standard mempools to generate transaction. 

There are few known situations.

1. Submitting transactions with invalid sequence numbers
2. Submitting transactions with invalid signatures
3. Creating a block where the total of Gas_wanted > max gas for a block.

If during block processing, the application track these events and slash the block producer.

The Slashing keeper needs a new method for slashing the block proposer. This should take a new parameter for malicious fees. This should result in both jailing the proposer and slashing the proposer for violations of the block production rules.

Issues 1 and 2 should be handled by adding a Slashing keeper to NewSigVerificationDecorator.

For issue 3, we need a transient store that totals the gas wanted for all the TXs in a block.

In the ante handler, we should track if the total gas wanted has exceeded the max gas per block and fail any txs after that and slash the proposer. This should be done with a new decorator.

### Slashing
The slashing parameter should slash once per infraction. The jail should remove the validator from the validator set for a fixed time that is not per infraction.

Governance should decide reasonable values for these parameters. 


## Consequences

This would introduce a new slashing that punishes validators for creating blocks that contain transactions that fail in the antehandler. This would be parameterized by a governance parameter that would slash proposers for the badly constructed blocks.




