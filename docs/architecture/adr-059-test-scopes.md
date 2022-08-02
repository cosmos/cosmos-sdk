# ADR 059: Test Scopes

## Changelog

* 2022-08-02: Initial Draft

## Status

PROPOSED Partially Implemented

## Abstract

Recent work in the SDK aimed at breaking apart the monolithic root go module has highlighted
shortcomings and inconsistencies in our testing paradigm.  This ADR clarifies a common
language for talking about test scopes and proposes an ideal state of tests at each scope.

## Context

[ADR-053: Go Module Refactoring](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-053-go-module-refactoring.md) expresses our desire for an SDK composed of many
independently versioned Go modules, and [ADR-057: App Wiring Part I](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-057-app-wiring-1.md) offers a methodology
for breaking apart inter-module dependencies through the use of dependency injection.  As
described in [EPIC: Separate all SDK modules into standalone go modules](https://github.com/cosmos/cosmos-sdk/issues/11899), module
dependencies are particularly complected in the test phase, where simapp itself is used as
the key test fixture in setting up and running tests.  It is clear that the successful
completion of Phases 3 and 4 in that EPIC require the resolution of this dependency problem.

In [EPIC: Unit Testing of Modules via Mocks](https://github.com/cosmos/cosmos-sdk/issues/12398) it was thought this Gordian knot could be
unwound by mocking all dependencies in the test phase for each module, but seeing how these
refactors were complete rewrites of test suites discussions began around the fate of the
existing integration tests.  One perspective is that they ought to be thrown out, another is
that integration tests have some utility of their own and a place in the SDK's testing story.

Another point of confusion has been the current state of CLI test suites, [x/auth](https://github.com/cosmos/cosmos-sdk/blob/0f7e56c6f9102cda0ca9aba5b6f091dbca976b5a/x/auth/client/testutil/suite.go#L44-L49) for
example.  In code these are called integration tests, but in reality function as end to end
tests by starting up a tendermint node and full application.  [EPIC: Rewrite and simplify
CLI tests](https://github.com/cosmos/cosmos-sdk/issues/12696) identifies the ideal state of CLI tests using mocks, but does not address the
place end to end tests may have in the SDK.

From here we identify three scopes of testing, **unit**, **integration**, **e2e** (end to
end), seek to define the boundaries of each, their shortcomings (real and imposed), and
ideal state in the SDK.

### Unit tests

Unit tests exercise the code contained in a single module (e.g. `/x/bank`) or package
(e.g. `/client`) in isolation from the rest of the code base.  Within this we identify two
levels of unit tests, *illustrative* and *journey*.  

Tests which exercise an atomic part of
a module in isolation - in this case we might do fixture setup/mocking of other parts of the
module tests which exercise a whole module's function with dependencies mocked, these are
almost like integration tests in that they exercise many things together but still use mocks

Example 1 [Bank keeper tests](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/x/bank/keeper/keeper_test.go#L94-L105) - A mock implementation of `AccountKeeper` is
supplied to the keeper constructor.

Example 2 [depinject tests](https://github.com/cosmos/cosmos-sdk/blob/main/depinject/provider_desc_test.go)

#### Limitations

Certain modules are tightly coupled beyond the test phase.  See the dependency report
for [bank->auth](./test-levels/bank-auth.txt), where 274 total usages of `x/auth` were found in `x/bank`, 50 of
which are in production code and 224 in test.  This tight coupling may suggest that
either the modules should be merged, or refactoring is required to abstract references
to the core types tying the modules together.  It could also indicate that these
modules should be tested together in integration tests beyond mocked unit tests.

In some cases setting up a test case for a module with many mocked dependencies can be quite
cumbersome and the resulting test may only show that the mocking framework works as expected
rather than working as a functional test of interdependent module behavior.

### Integration tests

Integration tests define and exercise relationships between an arbitrary number of modules
and/or application subsystems.

Wiring for integration tests is provided by `depinject` and some [helper code](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/testutil/sims/app_helpers.go#L95) starts up
a running application.  A section of the running application may then be tested.  Certain
inputs during different phases of the application life cycle are expected to produce
invariant outputs without too much concern for component internals.  This type of black box
testing has a larger scope than unit testing.

Example 1 [client/grpc_query_test/TestGRPCQuery](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/client/grpc_query_test.go#L111-L129) - This test is misplaced in `/client`,
but tests the life cycle of (at least) `runtime` and `bank` as they progress through
startup, genesis and query time.  It also exercises the fitness of the client and query
server without putting bytes on the wire through the use of [QueryServiceTestHelper](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/baseapp/grpcrouter_helpers.go#L31).

Example 2 `x/evidence` Keeper integration tests - Starts up an application composed of [8
modules](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/x/evidence/testutil/app.yaml#L1) with [5 keepers](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/x/evidence/keeper/keeper_test.go#L101-L106) used in the integration test suite.  One test in the suite
exercises [HandleEquivocationEvidence](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/x/evidence/keeper/infraction_test.go#L42) which contains many interactions with the staking
keeper.

#### Limitations

Setting up a particular input state may be more challenging since the application is
starting from a zero state.  Some of this may be addressed by good test fixture
abstractions with testing of their own.  Tests may also be more brittle, and larger
refactors could impact application initialization in unexpected ways with harder to
understand errors.  This could also be seen as a benefit, and indeed the SDK's current
integration tests were helpful in tracking down logic errors during earlier stages
of app-wiring refactors.

### Simulation / generative testing 

Simulations are a special case of integration tests where deterministically random module
operations are executed against a running simapp.  No assertions are made for the state
transitions resulting from module operations but any error will halt and fail the
simulation.

Modules must implement [AppModuleSimulation.WeightedOperations](https://github.com/cosmos/cosmos-sdk/blob/2bec9d2021918650d3938c3ab242f84289daef80/types/module/simulation.go#L31) to define their
simulation operations.  Note that not all modules implement this which may indicate a
gap in current simulation test coverage.

Modules not returning simulation operations: 

- `x/auth`
- `x/capability`
- `x/evidence`
- `x/mint`
- `x/params`

A separate binary, [runsim](https://github.com/cosmos/tools/tree/master/cmd/runsim), is responsible for kicking off some of these tests and
managing their life cycle.

#### Limitations

- [A success](https://github.com/cosmos/cosmos-sdk/runs/7606931983?check_suite_focus=true) may take a long time to run, 7-10 minutes per simulation in CI. 
- [Timeouts](https://github.com/cosmos/cosmos-sdk/runs/7606932295?check_suite_focus=true) sometimes occur on apparent successes without any indication why.
- Useful error messages not provided on [failure](https://github.com/cosmos/cosmos-sdk/runs/7606932548?check_suite_focus=true) from CI, requiring a developer to run
  the simulation locally to reproduce.

### End to end tests

End to end tests exercise the entire system as we understand it in as close an approximation
to a production environment as is practical.  Presently these tests are located at
[tests/e2e](https://github.com/cosmos/cosmos-sdk/tree/main/tests/e2e) and rely on [testutil/network](https://github.com/cosmos/cosmos-sdk/tree/main/testutil/network) to start up an in-process Tendermint node.

#### Limitations

In general the limitations of end to end tests are orchestration and compute cost.
Scaffolding is required to start up and run a prod-like environment and the this
process takes much longer to start and run than unit or integration tests.

Global locks present in Tendermint code cause stateful starting/stopping to sometimes hang
or fail intermittently when run in a CI environment.

The scope of e2e tests has been complected with command line interface testing.

## Decision


## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
