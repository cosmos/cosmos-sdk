## Software Upgrade

Software is upgraded according to a 4-step process.

1. Agree on software upgrade plan via governance (currently via a plaintext proposal).
2. Once the upgrade plan is accepted, validators each update their software according to proposal from step 1.
3. Each upgraded validator submits a transaction to the SoftwareUpgradeKeeper.
4. Once a sufficient quorum of validators has completed step 3, new logic kicks
   in in the form of a special parameter.

### Step 1: Agree on the Upgrade Plan

First, an English plaintext proposal should be submitted based on the following template:

```
# SOFTWARE UPGRADE PROPOSAL

This is a proposal to upgrade the software.

## Outline:

* Current version: [XXX]
* Current version commit hash: [XXX]
* New version Github URL: https://github.com/...
* New version: [XXX]
* New version commit hash: [XXX]
* Upgrade prepared by:
 - <name/handle> on <date>; commit signature: https://github.com/...
* Audited by:
 - <name/handle> on <date>; report: https://github.com/...
 - <name/handle> on <date>; report: https://github.com/...

## Description:

Describe the purpose of the upgrade and other details.

- [ ] Link to any previous discussions
- [ ] Link to any alternative proposals

## Requirements:

Describe any requirements

- [ ] Any new hardware requirements?
- [ ] Any new bandwidth requirements?
- [ ] Any dependencies on other proposals?

## Procedure:

* Deadline: By 2 weeks from after the proposal has been accepted.

Once this proposal has been accepted, validators should upgrade to the new
software as described above by the deadline.  Once the new software is
deployed, the new software will automaticaly submit a message to the
SoftwareUpgradeKeeper.  The SoftwareUpgradeKeeper ensures that a sufficient
quorum of validators (SOFTWARE_UPGRADE_QUORUM parameter) is running a specific
version of the blockchain software.

Once a sufficient quorum has been reached, the "SOFTWARE_UPGRADE_[XXX]_ENABLED"
parameter will get set to true, thus switching on the new logic.

## Conflict Resolution

In case there are conflicting proposals, the most recent proposal to get
accepted should be followed.  In other words, if another proposal gets accepted
after this proposal does which conflicts with this proposal, the new proposal
should be heeded instead of the process declared here.
```

It is recommended that all links point to Github for now, for accountability.

This proposal template assumes that much discussion about the upgrade has
already taken place elsewhere.  To gauge interest in a proposed upgrade,
users should instead submit a survey proposal with the following template:

Notethat the `Procedure` section has already been filled out.  Do not change
this section unless there is a good reason to do so.

```
# SOFTWARE UPGRADE SURVEY

THIS IS A NONBINDING SURVEY.  VOTES HERE ARE ONLY INTENDED TO GUAGE INTEREST.
IF THIS PROPOSAL MANDATES ANY ACTION, IT IS THE VOTER'S DUTY TO VOTE "NO".

<fill out as much as possible from the SOFTWARE UPGRADE PROPOSAL template>
```

### Step 2: Validators upgrade their software

Once the proposal from step 1 has been accepted, validators should upgrade
their software.  For the standard upgrade procedure, it is OK for non-validator
full nodes to upgrade as well.

### Step 3: Update the SoftwareUpgradeKeeper

The new version of the software should automatically include a (single) signed
message to the SoftwareUpgradeKeeper w/ the new version of software that the
validator is running.  Even if a validator were to change their software
version multiple times, the SoftwareUpgradeKeeper ensures that nothing happens
until a quorum (SOFTWARE_UPGRADE_QUORUM) are on an identical version of the
software.

Only validators can vote here, as delegators do not actually run the validation
software.  Delegators should make sure to vote NO or veto objectionable
software upgrade proposals in the first step of this process.

### Step 4: Auto-switch to new logic

Once a quorum has been reached, the SoftwareUpgradeKeeper sets a new parameter
SOFTWARE_UPGRADE_[XXX]_ENABLED=true, where XXX is the commit hash of the new
software.

### Contingincies

TODO: What happens when there are two conflicting software upgrade proposals?
Currently the SoftwareUpgradeKeeper ensures that only one upgrade happens at a
time, but the only protocol that ensures accountability in the face of
conflicting software upgrade proposals is the "Conflict Resolution" section.
We should have a constitution that acts as a human-level protocol spec that
defines how to deal with conflicting proposals in general.

TODO: Document more edge cases.

### Future improvement

Any convention/template that develops in the plaintext proposal could be
codified as a "SoftwareUpgradeProposal", which more or less acts as the
plaintext proposal does here in step 1.  The special categorization signals to
the greater ecosystem (e.g. non-validator full node operators) that an upgrade
is being considered.

In the future, the Cosmos Hub can include logic for hosting proposal
discussions even before they reach the deposit requirement.
