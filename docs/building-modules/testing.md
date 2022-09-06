<!--
order: 15
-->

# Module & App Testing

The Cosmos SDK contains different types of [tests](https://martinfowler.com/articles/practical-test-pyramid.html).
These tests have different goals and are used at different stages of the development cycle.
We advice, as a general rule, to use tests at all stages of the development cycle.
It is adviced, as a chain developer, to test your application and modules in a similar way than the SDK.

The rationale behind testing can be found in [ADR-59](https://docs.cosmos.network/main/architecture/adr-059-test-scopes.html).

## Unit Tests

Unit tests are the lowest test category of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
All packages and modules should have unit test coverage. Modules should have their dependencies mocked: this means mocking keepers for example.

+++ https://github.com/cosmos/cosmos-sdk/blob/a92c291880eb6240b7221173282fee0c5f2adb05/x/gov/keeper/keeper_test.go#L21-L35

The SDK uses `mockgen` to generate mocks for keepers:

+++ https://github.com/cosmos/cosmos-sdk/blob/dd556936b23d7443cb7fb1da394c35117efa9da7/scripts/mockgen.sh#L12-L29

## Integration Tests

Integration tests are at the second level of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
In the SDK, we locate our integration tests under [`/tests/integrations`](https://github.com/cosmos/cosmos-sdk/tree/main/tests/integration).

The goal of these integration tests is to test a component with a minimal application (i.e. not `simapp`). The minimal application is defined with the help of `depinject` – the sdk dependency injection framework, and includes all necessary modules to test the component:

+++ https://github.com/cosmos/cosmos-sdk/blob/e09516f4795c637ab12b30bf732ce5d86da78424/tests/integration/bank/keeper/keeper_test.go#L188-L199

## Simulations

Simulations uses has well a minimal application, built with `depinject`:

Following is an example for `x/gov/` simulations:

+++ https://github.com/cosmos/cosmos-sdk/blob/0fbcb0b18381d19b7e556ed07e5467129678d68d/x/gov/simulation/operations_test.go#L290-L307

+++ https://github.com/cosmos/cosmos-sdk/blob/main/x/gov/simulation/operations_test.go#L67-L109

## End-to-end Tests

End-to-end tests are at the top of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
They must test the whole application flow, from the user perspective (for instance, CLI tests). They are located under [`/tests/e2e`](https://github.com/cosmos/cosmos-sdk/tree/main/tests/e2e).

For that, the SDK is using `simapp` but you should use your own application (`appd`).
Here are some examples:

* SDK E2E tests: <https://github.com/cosmos/cosmos-sdk/tree/main/tests/e2e>.
* Cosmos Hub E2E tests: <https://github.com/cosmos/gaia/tree/main/tests/e2e>.
* Osmosis E2E tests: <https://github.com/osmosis-labs/osmosis/tree/main/tests/e2e>.

## Summary

| Scope       | App Fixture | Mocks? |
| ----------- | ----------- | ------ |
| Unit        | None        | Yes    |
| Integration | `depinject` | Some   |
| Simulation  | `depinject` | No     |
| E2E         | `appd`      | No     |
