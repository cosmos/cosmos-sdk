# ADR 039: Epoched Staking

## Changelog

- 10-Feb-2021: Initial Draft

## Authors

- Dev Ojha (@valardragon)
- Sunny Aggarwall (@sunnya97)

## Status

Proposed

## Abstract

This ADR updates the proof of stake module to buffer the staking weight updates for a number of blocks before updating the consensus' staking weights. The length of the buffer is dubbed an epoch. The prior functionality of the staking module is then a special case of the abstracted module, with the epoch being set to 1 block.

## Context

The current proof of stake module takes the design decision to apply staking weight changes to the consensus engine immediately. This means that delegations and unbonds get applied immediately to the validator set. This decision was primarily done as it was implementationally simplest, and because we at the time believed that this would lead to better UX for clients.

An alternative design choice is to allow buffering staking updates (delegations, unbonds, validators joining) for a number of blocks. This 'epoch'd proof of stake consensus provides the guarantee that the consensus weights for validators will not change mid-epoch, except in the event of a slash condition. 

The decision to have immediate execution of staking changes was primarily done as it was implementationally simplest, and because we at the time believed that this would lead to better UX for clients. The UX hurdle may not be as significant as was previously thought, since it is possible to provide users acknowledgement that their bond was recorded and will be executed.

Furthermore, it has become clearer over time that immediate execution of staking events comes with limitations, such as:

* Threshold based cryptography. One of the main limitations is that because the validator set can change so regularly, it makes the running of multiparty computation by a fixed validator set difficult. Many threshold-based cryptographic features for blockchains such as randomness beacons and threshold decryption require a computationally-expensive DKG process (will take much longer than 1 block to create). To productively use these, we need to guarantee that the result of the DKG will be used for a reasonably long time. It wouldn't be feasible to rerun the DKG every block. By epoching staking, it guarantees we'll only need to run a new DKG once every epoch.

* Light client efficiency. This would lessen the overhead for IBC. Because of the lite client bisection algorithm, the number of headers you need to verify is related to bounding the validator set diffs between two successively verified headers. By limiting the frequency of validator set changes, we can reduce the size of IBC lite client proofs.

* Fairness of deterministic leader election. Currently we have no ways of reasoning of fairness of deterministic leader election in the presence of staking changes without epochs (tendermint/spec#217). Adding epochs at least makes it easier for our deterministic leader election to match something we can prove secure. (Albeit, we still haven’t proven if our current algorithm is fair with > 2 validators)

* Staking derivative design. Currently, reward distribution is done lazily using the F1 fee distribution. While saving computational complexity, lazy accounting increases “statefulness of staking”. Right now, each delegation entry has to track the time of last withdrawal. Handling this can be a challenge for some staking derivatives designs (see example). Force-withdrawing rewards to users can help solve this, however it is infeasible to force-withdraw rewards to users on a per block basis. With epoching, a chain could more easily alter the design to have rewards be forcefully withdrawn (iterating over delegator accounts only once per-epoch), and thus remove the time of delegation from state. This preliminarily seems like it may be of utility in certain staking derivative designs.

## Design considerations

### Slashing

There is a design consideration for whether to apply a slash immediately or at the end of an epoch. A slash event should apply to only members who are actually staked during the time of the infraction, namely during the epoch the slash event occured.

Applying it immediately can be viewed as offering greater consensus layer security, at potential costs to the aforementioned usecases. The benefits of immediate slashing for consensus layer security can be all be obtained by executing the validator jailing immediately (thus removing it from the validator set), and delaying the actual slash change to the validator's weight until the epoch boundary. For the use cases mentioned above, workarounds can be integrated to avoid problems, as follows:

- For threshold based cryptography, it can keep using the original keep epoch weights for the cryptography thresholds, but allow the underlying finality to benefit from extra security more quickly.
- For light client efficiency, there can be a bit included in the header indicating an intra-epoch slash (ala https://github.com/tendermint/spec/issues/199).
- For fairness of deterministic leader election, this will cause problems with the formalization of it / proximity of implementation to formally provable spec. However, a less formal claim can be made that the amount lost due to the slash should hopefully outweigh slight biases into the leader election process. This claim is dubious with the presence of MEV, but potentially formal upperbounds on MEV for fairness here could be derived.
- For staking derivative design, this will not cause problems with the suggested design there, nor does it increase the stateful of staking. (As whether a slash has occured is fully queryable given the validator address)

However, for achieving consensus layer security, it suffices to apply the validator jailing immediately, but still delay the actual slash changes to waiting until the end of the epoch. This largely mitigates the concern for the fairness of deterministic leader election as well, since that validator is removed the set being rotated from immediately.

### Token lockup

When someone makes a transaction to delegate, even though they are not immediately staked, their tokens should be moved into a pool managed by the staking module which will then be used at the end of an epoch. This prevents concerns where they stake, and then spend those tokens not realizing they were already allocated for staking, and thus having their staking tx fail.

### Pipelining the epochs

For threshold based cryptography in particular, we need a pipeline for epoch changes. This is because when we are in epoch N, we want the epoch N+1 weights to be fixed so that the validator set can do the DKG accordingly. So if we are currently in epoch N, the stake weights for epoch N+1 should already be fixed, and new stake changes should be getting applied to epoch N + 2.

This can be handled by making a parameter for the epoch pipeline. This parameter should not be alterable except during hard forks, to mitigate implementation complexity of switching the pipeline length.

### Rewards

Even though all staking updates are applied at epoch boundaries, rewards can still be distributed immediately when they are claimed. This is because they do not affect the current stake weights, as we do not implement auto-bonding of rewards. If such a feature were to be implemented, it would have to be setup so that rewards are auto-bonded at the epoch boundary.

## Decision

__Step-1__:  Implement buffering of all staking and slashing messages. 

First we create a pool for storing tokens that are being bonded, but should be applied at the epoch boundary called the `EpochDelegationPool`. Then, we have two separate queues, one for staking, one for slashing. We describe what happens on each message being delivered below:

### Staking messages
- **MsgCreateValidator**: Move user's self-bond to `EpochDelegationPool` immediately. Queue a message for the epoch boundary to handle the self-bond, taking the funds from the `EpochDelegationPool`. If Epoch execution fail, return back funds from `EpochDelegationPool` to user's account.
- **MsgEditValidator**: Validate message and if valid queue the message for execution at the end of the Epoch.
- **MsgDelegate**: Move user's funds to `EpochDelegationPool` immediately. Queue a message for the epoch boundary to handle the delegation, taking the funds from the `EpochDelegationPool`. If Epoch execution fail, return back funds from `EpochDelegationPool` to user's account.
- **MsgBeginRedelegate**: Validate message and if valid queue the message for execution at the end of the Epoch.
- **MsgUndelegate**: Validate message and if valid queue the message for execution at the end of the Epoch.

### Slashing messages
- **MsgUnjail**: Validate message and if valid queue the message for execution at the end of the Epoch.
- **Slash Event**: Whenever a slash event is created, it gets queued in the slashing module to apply at the end of the epoch. The queues should be setup such that this slash applies immediately.

### Evidence Messages
 - **MsgSubmitEvidence**: This gets executed immediately, and the validator gets jailed immediately. However in slashing, the actual slash event gets queued.

Then we add methods to the end blockers, to ensure that at the epoch boundary the queues are cleared and delegation updates are applied.


__Step-2__: Implement querying of queued staking txs. 

When querying the staking activity of a given address, the status should return not only the amount of tokens staked, but also if there are any queued stake events for that address. This will require nodes supporting querying to either do some more indexing to have this be efficiently queryable, or to have transactions 

__Step-3__: Adjust gas

Currently gas represents the cost of executing a transaction when its done immediately. (Merging together costs of p2p overhead, state access overhead, and computational overhead) However, now a transaction can cause computation in a future block, namely at the epoch boundary.

To handle this, we should initially include parameters for estimating the amount of future computation (denominated in gas), and add that as a flat charge needed for the message.
We leave it as out of scope for how to weight future computation versus current computation in gas pricing, and have it set such that the are weighted equally for now.

## Consequences

### Positive

* Abstracts the proof of stake module that allows retaining the existing functionality
* Enables new features such as validator-set based threshold cryptography

### Negative

* Increases complexity of integrating more complex gas pricing mechanisms, as they now have to consider future execution costs as well.
