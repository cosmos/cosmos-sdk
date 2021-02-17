<!--
order: 3
-->

# State Transitions

## Client State Verification Functions

Successful state verification by a solo machine light client will result in:

- the sequence being incremented by 1.

## Update By Header

A successful update of a solo machine light client by a header will result in:

- the public key being updated to the new public key provided by the header. 
- the diversifier being updated to the new diviersifier provided by the header.
- the timestamp being updated to the new timestamp provided by the header.
- the sequence being incremented by 1
- the consensus state being updated (consensus state stores the public key, diversifier, and timestamp)

## Update By Governance Proposal

A successful update of a solo machine light client by a governance proposal will result in:

- the client state being updated to the substitute client state
- the consensus state being updated to the substitute consensus state (consensus state stores the public key, diversifier, and timestamp)
- the frozen sequence being set to zero (client is unfrozen if it was previously frozen).

## Upgrade

Client udgrades are not supported for the solo machine light client. No state transition occurs.

## Misbehaviour

Successful misbehaviour processing of a solo machine light client will result in:

- the frozen sequence being set to the sequence the misbehaviour occurred at
