
# ADR 047: Governance Voting Profiles

## Changelog

- {date}: Initial Version.

## Status

DRAFT - Not Implemented

## Abstract

The Governance Voting Profiles are designed to provide an alternative set of voting configuration that can be
optioned when a proposal is submitted. The main driver for profiles are the need for different procedures for different
situations.  The largest contrast in requirements occurs between standard governance proposals and the need for 
an expedited process.

## Context

During the normal operation of a blockchain network there are occasionally instances where a time sensitive decision must 
be made.  These decisions are often in response to an existing vulnerability in the network that when disclosed places 
the network at additional risk.

An expedited voting option is an important tool for responding to time sensitive situations where a normal full length
voting process introduces additional risk to the network. A recent (SEP2021) example of this situation occurred with
the [Compound Protocol][1] where an [identified][2] [issue][3] which was quickly resolved and submitted as a
[governance proposal][4].  During the governance voting period (7 days) the network was vulnerable and millions of USD
in value were withdrawn due to the exploit. 

Recognizing that different timelines for different decisions make sense this architecture decision record outlines
changes to support multiple voting configurations as needed for a given network.

### Existing Voting Considerations

The existing governance voting period is configured with constraints designed to promote stability and high levels of
stakeholder participation.  The minimum deposit levels are set at levels to minimize spam without creating a barrier to
stakeholders using the process.  Too short of a deposit window or voting period would limit the number of stakeholders
that see the proposal as well the period for education or outreach from stakeholders to campaign for or against the 
proposal.

### Expedited Voting Considerations

An expedited voting process is focused on reaching a consensus in the shortest time possible.  It is expected that a 
certain amount of offline coordination will occur prior to the proposal being published such that when posted the 
stakeholders can respond quickly to the measure.  In the interest of getting a decision quickly a proposal should be
passed as soon as the required threshold is met.  If a governance proposal can pass with variable endpoint then it may
be desirable to have a method to provide the effective blockheight to the proposals being passed.

### Exigency and Network Governance

The initial example of the Compound Protocol issue illustrates the need for a quick response to a known issue.  In this 
context the flaw is identified and as quick a resolution as possible once the public announcement is made is desired.
Unlike a normal voting process a higher level of participation is not the goal, the primary stakeholders that stand to 
lose funds should be prepared to vote ahead of the proposal submission.  The deposit window should be considered largely
irrelevant as a situation with a grave risk should allow commensurate amounts of capital to be posted in bond.


## Decision

Given the needs of governance actions for standard proposals differ greatly from those of expedited proposals we will 
modify the governance module to support multiple voting period and proposal configurations.  The existing parameters
for voting and proposal creation will be migrated into a default profile.  Networks will have the option to add
additional profiles (such as an exigent process) according to their needs.

Given the need for a method to respond to network crises through governance we will create an expedited voting process
that mirrors the existing process combined with a separate set of tuning parameters.  We will craft this voting process
to end as soon as one of the conditions to pass, fail, or veto the measure is met.

### Existing Governance Configuration

The existing governance process uses the following structures to hold configuration parameters.

```proto
// DepositParams defines the params for deposits on governance proposals.
message DepositParams {
  //  Minimum deposit for a proposal to enter voting period.
  repeated cosmos.base.v1beta1.Coin min_deposit = 1;
  //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2  months.
  google.protobuf.Duration max_deposit_period = 2;
}

// VotingParams defines the params for voting on governance proposals.
message VotingParams {
  //  Length of the voting period.
  google.protobuf.Duration voting_period = 1;
}

// TallyParams defines the params for tallying votes on governance proposals.
message TallyParams {
  //  Minimum percentage of total stake needed to vote for a result to be considered valid.
  bytes quorum = 1;
  //  Minimum proportion of Yes votes for proposal to pass. Default value: 0.5.
  bytes threshold = 2;
  //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Default value: 1/3.
  bytes veto_threshold = 3;
}
```

The updated method will require additional copies of this configuration.  We will build a new structure to hold the
existing configuration which will allow multiple copies to be maintained using instances of this new params structure.
These copies will be maintained as an array of configurations with a single profile annotated as the default.

```proto
// Voting Configuration

// VotingParams defines the params for voting on governance proposals.
message VotingParams {
  //  Length of the voting period.
  google.protobuf.Duration voting_period = 1;
  google.protobuf.Duration minimum_voting_period = 2;

  cosmos.gov.v1beta1.DepositParams deposit = 3;
  cosmos.gov.v1beta1.TallyParams tally = 4;
  bytes voting_process_name = 4;
}

```

## Consequences

Adding the ability to configure multiple types of voting periods that are suitable for different types of situations
will allow networks to customize a set of processes according to their needs.  This change captures the existing voting
parameters into a new default profile preserving existing behavior.  It will allow a network to add additional sets of 
governance parameters for situations such as the ability to respond to a network emergency.


### Backwards Compatibility

The existing governance voting and deposit parameters will be migrated into a default profile.  Any proposals submitted
that do not specify an alternate profile will use this default.  This configuration will allow the current behavior to
be maintained for all networks.  Networks that wish to support additional profiles can add these according to their 
unique needs.

### Positive

- An alternative set of voting parameters is available for chains to customize for an alternate process
- Existing proposal configurations that encourage participation can be maintained
- Future proposals that may warrant a longer than usual voting period for whatever reason could be supported
- No new configurations or processes are added to networks that do not wish to use these capabilities

### Negative

- Accelerated governance votes increase risk to networks by potentially reducing participation and visibility of proposals.
- Increased configuration complexity for the governance module
- Existing documentation and examples will require updates

### Neutral

- None

## Further Discussions

- [Cosmos SDK #10014: Expedited Governance Proposal Option](https://github.com/cosmos/cosmos-sdk/issues/10014)


## References

- [1]: https://compound.finance
- [2]: https://cointelegraph.com/news/compound-supply-bug-mistakenly-rewarded-users-with-70m-in-tokens
- [3]: https://twitter.com/Mudit__Gupta/status/1443454935639609345?ref_src=twsrc%5Etfw%7Ctwcamp%5Etweetembed%7Ctwterm%5E1443454938151985153%7Ctwgr%5E%7Ctwcon%5Es2_ref_url=https%3A%2F%2Fcointelegraph.com%2Fnews%2Fcompound-supply-bug-mistakenly-rewarded-users-with-70m-in-tokens
- [4]: https://compound.finance/governance/proposals/64