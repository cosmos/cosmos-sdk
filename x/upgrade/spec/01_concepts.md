<!--
order: 1
-->

# Concepts

## Plan

The `x/upgrade` module defines a `Plan` type in which a live upgrade is scheduled
to occur. A `Plan` can be scheduled at a specific block height or time, but not both.
A `Plan` is created once a (frozen) release candidate along with an appropriate upgrade
`Handler` (see below) is agreed upon, where the `Name` of a `Plan` corresponds to a
specific `Handler`. Typically, a `Plan` is created through a governance proposal
process, where if voted upon and passed, will be scheduled. The `Info` of a `Plan`
may contain various metadata about the upgrade, typically application specific
upgrade info to be included on-chain such as a git commit that validators could
automatically upgrade to.

### Sidecar Process

If an operator running the application binary also runs a sidecar process to assist
in the automatic download and upgrade of a binary, the `Info` allows this process to
be seamless. Namely, the `x/upgrade` module fulfills the
[cosmosd Upgradeable Binary Specification](https://github.com/regen-network/cosmosd#upgradeable-binary-specification)
specification and `cosmosd` can optionally be used to fully automate the upgrade
process for node operators. By populating the `Info` field with the necessary information,
binaries can automatically be downloaded. See [here](https://github.com/regen-network/cosmosd#auto-download)
for more info.

```go
type Plan struct {
  Name   string
  Time   Time
  Height int64
  Info   string
}
```

## Handler

The `x/upgrade` module facilitates upgrading from major version X to major version Y. To
accomplish this, node operators must first upgrade their current binary to a new
binary that has a corresponding `Handler` for the new version Y. It is assumed that
this version has fully been tested and approved by the community at large. This
`Handler` defines what state migrations need to occur before the new binary Y
can successfully run the chain. Naturally, this `Handler` is application specific
and not defined on a per-module basis. Registering a `Handler` is done via
`Keeper#SetUpgradeHandler` in the application.

```go
type UpgradeHandler func(Context, Plan)
```

During each `EndBlock` execution, the `x/upgrade` module checks if there exists a
`Plan` that should execute (is scheduled at that time or height). If so, the corresponding
`Handler` is executed. If the `Plan` is expected to execute but no `Handler` is registered
or if the binary was upgraded too early, the node will gracefully panic and exit.

## Proposal

Typically, a `Plan` is proposed and submitted through governance via a `SoftwareUpgradeProposal`.
This proposal prescribes to the standard governance process. If the proposal passes,
the `Plan`, which targets a specific `Handler`, is persisted and scheduled. The
upgrade can be delayed or hastened by updating the `Plan.Time` in a new proposal.

```go
type SoftwareUpgradeProposal struct {
  Title       string
  Description string
  Plan        Plan
}
```

### Cancelling Upgrade Proposals

Upgrade proposals can be cancelled. There exists a `CancelSoftwareUpgrade` proposal
type, which can be voted on and passed and will remove the scheduled upgrade `Plan`.
Of course this requires that the upgrade was known to be a bad idea well before the
upgrade itself, to allow time for a vote.

If such a possibility is desired, the upgrade height is to be
`2 * (VotingPeriod + DepositPeriod) + (SafetyDelta)` from the beginning of the
upgrade proposal. The `SafetyDelta` is the time available from the success of an
upgrade proposal and the realization it was a bad idea (due to external social consensus).

A `CancelSoftwareUpgrade` proposal can also be made while the original
`SoftwareUpgradeProposal` is still being voted upon, as long as the `VotingPeriod`
ends after the `SoftwareUpgradeProposal`.
