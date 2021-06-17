<!--
order: 13
-->

# Errors

This document outlines the recommended usage and APIs for error handling in Cosmos SDK modules. {synopsis}

Modules are encouraged to define and register their own errors to provide better
context on failed message or handler execution. Typically, these errors should be
common or general errors which can be further wrapped to provide additional specific
execution context.

## Registration

Modules should define and register their custom errors in `x/{module}/types/errors.go`. Registration
of errors is handled via the `types/errors` package.

Example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.38.1/x/distribution/types/errors.go#L1-L21

Each custom module error must provide the codespace, which is typically the module name
(e.g. "distribution") and is unique per module, and a uint32 code. Together, the codespace and code
provide a globally unique SDK error. Typically, the code is monotonically increasing but does not
necessarily have to be. The only restrictions on error codes are the following:

* Must be greater than one, as a code value of one is reserved for internal errors.
* Must be unique within the module.

Note, the SDK provides a core set of *common* errors. These errors are defined in `types/errors/errors.go`.

## Wrapping

The custom module errors can be returned as their concrete type as they already fulfill the `error`
interface. However, module errors can be wrapped to provide further context and meaning to failed
execution.

Example:

+++ https://github.com/cosmos/cosmos-sdk/blob/b2d48a9e815fe534a7faeec6ca2adb0874252b81/x/bank/keeper/keeper.go#L85-L122

Regardless if an error is wrapped or not, the SDK's `errors` package provides an API to determine if
an error is of a particular kind via `Is`.

## ABCI

If a module error is registered, the SDK `errors` package allows ABCI information to be extracted
through the `ABCIInfo` API. The package also provides `ResponseCheckTx` and `ResponseDeliverTx` as
auxiliary APIs to automatically get `CheckTx` and `DeliverTx` responses from an error.
