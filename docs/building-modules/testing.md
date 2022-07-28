<!--
order: 15
-->

# Module & App Testing

The Cosmos SDK contains different types of tests. The following are some of the types of tests that are available and their goals.
It is adviced, as a chain developer, to test your application and module in a similar way.

## End-to-End Tests

* Where: [`tests/e2e`](https://github.com/cosmos/cosmos-sdk/tree/main/tests/e2e).

End-to-End tests are tests that run against the chain (`SimApp`). They are useful for testing the full functionality of a module against the application.
As a chain developer, you can test your own chain against the end-to-end tests of the SDK.

## Unit Tests

* Where: `x/{mod}/keeper`

The SDK uses mocking for unit testing. It allows us to test the behavior of a module without the need to run the full chain.
The expected keepers of a module are mocked:

+++ https://github.com/cosmos/cosmos-sdk/blob/main/x/auth/testutil/expected_keepers_mocks.go

## Simulations

* Where: `x/{mod}/simulation`

Simulation tests are using `AppConfig`, which allows us to test a module without depending on the full chain and all its modules.

## CLI Tests

https://github.com/cosmos/cosmos-sdk/issues/12696

TBD.
