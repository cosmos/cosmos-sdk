<!--
order: 5
-->

# Governance Proposals

In uncommon situations, a highly valued client may become frozen due to uncontrollable 
circumstances. For example, the chain a light client represents might fork into two
new chains. The light client will now have two valid, but conflicting headers at the same
height. Evidence of such misbehaviour is likely to be submitted resulting in a frozen light
client. 

Frozen light clients cannot be updated under any circumstance except via a governance proposal.
Since validators can arbitarily agree to make state transitions that defy the written code, a 
governance proposal has been added to ease the complexity of unfreezing or updating clients
which have become "stuck". Unfreezing clients, re-enables all of the channels built upon that
client. This may result in recovery of otherwise lost funds. 

In the case that a highly valued light client is frozen, expired, or rendered non-updateable, a
governance proposal may be submitted to update this client, known as the subject client. The 
proposal includes the client identifier for the subject, the client identifier for a substitute
client, and an initial height to reference the substitute client from. Light client implementations
may implement custom updating logic, but in most cases, the subject will be updated with information
from the substitute client, if the proposal passes. The substitute client is used as a "stand in"
while the subject is on trial. It is best practice to create a substitute client *after* the subject
has become frozen to avoid the substitute from also becoming frozen. An active substitute client
allows headers to be submitted during the voting period to prevent accidental expiry once the proposal
passes. 

Example:

There exists a very common client called "ethereum-0" which is a light client for the Ethereum chain.
Ethereum undergoes a fork, creating two valid headers. As a result, misbehaviour evidence is submitted 
to "ethereum-0" rendering it frozen. The Cosmos Hub decides that "Ethereum 2.0" fork is the desired 
counterparty chain for this client. The proposer of the governance proposal to unfreeze "ethereum-0"
creates a new "ethereum-{N}" client with the exact same parameters (except for latest height, 
frozen height, and chain-id). Since the substitute client was created *after* the fork, there are
no conflicting headers to freeze the client. During the voting period, the substitute client can
constantly be updated to prevent the subject from being expired at the end of the voting period. 
If the vote passed, the "ethereum-0" client would have an updated chain-id and consensus states copied 
directly from the substitute client's store. 
