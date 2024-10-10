**WARNING: This is an API preview! Most code won't work or even type check properly!**
This crate defines types and macros for constructing easy to use account and module implementations.
It integrates with the encoding layer but does not specify a state management framework.

## Context

The [`Context`] struct gets passed to all handler functions that interact with state but
what exactly is it? In the Golang [Cosmos SDK](https://github.com/cosmos/cosmos-sdk),
`Context` was never precisely defined and ended up becoming sort of a bag of variables
which were passed around everywhere. This made it difficult to understand what was
actually being passed around and what was being used.

[`Context`] here is defined to have the following purposes:
* wrap some basic information about the current call, but not the full message header
* hold a handle to [`ixc_schema`]'s bump allocator
* hold a handle to the function used to send messages
