<!--
order: 15
-->

# Module & App Testing

The Cosmos SDK contains different types of [tests](https://martinfowler.com/articles/practical-test-pyramid.html). The following are some of the types of tests that are available and their goals.
It is adviced, as a chain developer, to test your application and module in a similar way.

The rationale behind testing can be found in [ADR-59](https://docs.cosmos.network/main/architecture/adr-059-test-scopes.html).

## Unit Tests

Unit tests are the lowest of the [test pyramid](https://martinfowler.com/articles/practical-test-pyramid.html).
All modules should have mocked unit test coverage. This means mocking keepers or external dependencies.
The SDK uses `mockgen` to generate mocks for keepers:

+++ https://github.com/cosmos/cosmos-sdk/blob/dd556936b23d7443cb7fb1da394c35117efa9da7/scripts/mockgen.sh#L12-L29

+++ https://github.com/cosmos/cosmos-sdk/blob/a92c291880eb6240b7221173282fee0c5f2adb05/x/gov/keeper/keeper_test.go#L21-L35

## Integration Tests

## Simulations

## End-to-end Tests

## Summary

| Scope       | App Fixture | Mocks? |
| ----------- | ----------- | ------ |
| Unit        | None        | Yes    |
| Integration | `depinject` | Some   |
| Simulation  | simapp      | No     |
| E2E         | simapp      | No     |
