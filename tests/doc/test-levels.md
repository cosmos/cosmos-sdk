# Test scopes

## Overview

This document is an analysis of testing scopes, current and proposed, in the SDK as they
relate to [EPIC: Separate all SDK modules into standalone go modules](https://github.com/cosmos/cosmos-sdk/issues/11899).  Some tests have
been written such that they introduce additional dependencies for the test phase only, and
these must be addressed for the successful completion of Phases 3 and 4.

## Unit tests

Unit tests exercise the code contained in a single module (e.g. `/x/bank`) or package
(e.g. `/client`) in isolation from the rest of the code base.  If an SDK component has
dependencies on other modules they must be mocked.

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

Setting up a test case for a module with many mocked dependencies is some cases is quite cumbersome
and the resulting test may only show that the mocking framework works as expected rather than
working as a functional test of interdependent module behavior.

#### Proposal

Continue to rewrite existing integration tests into mocked unit tests as per [EPIC: Unit testing of
modules via mocks](https://github.com/cosmos/cosmos-sdk/issues/12398) but maintain current integration tests, moving them to `/tests/integration`.

All modules must have mocked unit test coverage.

Unit tests outnumber integration tests.

Unit tests must not introduce additional dependencies beyond those already existing in
production code.

## Integration tests

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
starting from "first principles".  Some of this may be addressed by good test fixture
abstractions with testing of their own.  Tests may also be more brittle, and larger
refactors could impact application initialization in unexpected ways with harder to
understand errors.  This could also be seen as a benefit, and indeed the SDK's current
integration tests were helpful in tracking down logic errors during earlier stages
of app-wiring refactors.

#### Proposal

Maintain existing integration tests, but move them to `/tests/integration`.

To help limit scope and complexity, prefer the smallest possible number of modules in
application startup possible, i.e. don't depend on simapp.

Maintain a *small* number of integration tests which use simapp directly to start up
and test application operations.

Integration tests outnumber end to end tests.

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

### Limitations

- [A success](https://github.com/cosmos/cosmos-sdk/runs/7606931983?check_suite_focus=true) may take a long time to run, 7-10 minutes per simulation in CI. 
- [Timeouts](https://github.com/cosmos/cosmos-sdk/runs/7606932295?check_suite_focus=true) sometimes occur on apparently successes without any indication why.
- Useful error messages not provided on [failure](https://github.com/cosmos/cosmos-sdk/runs/7606932548?check_suite_focus=true) from CI, requiring a developer to run
  the simulation locally to reproduce.

#### Proposal

No changes proposed at this time.

## End to end tests

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

#### Proposal

Transition existing end to end tests integration tests by removing the dependency on the
test network and in-process tendermint node, ensuring we're not losing test coverage. 

Keep a small number existing end to end tests but transition to a runner powered by Docker
via [dockertest](https://github.com/ory/dockertest).

Introduce end to end tests exercising a full network upgrade.

Introduce true CLI unit tests through Tendermint mocking as demonstrated in [PR#12706](https://github.com/cosmos/cosmos-sdk/pull/12706).
This should be repeated for all module CLI tests.
