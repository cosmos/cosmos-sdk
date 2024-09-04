# RFC {RFC-NUMBER}: {TITLE}

## Changelog

* 2024-09-04: Initial draft

## Background

[RFC 002](./rfc-003-crosslang.md) introduces a specification for cross-language
message passing. One large follow-up item is defining APIs for state access.
This RFC addresses that.

## Proposal

We define the following APIs in terms of the cross-language **message packet** and error specification. Each message name is a module message name and internally must be prefixed as such.

The following APIs are expected to be used by 

### State API

The state API defines the following messages:

#### `cosmos.state.v1.new_branch`

Takes no parameters and returns a new volatile state token which branches/caches off the current state token as **output parameter 1**. Returns an error if the current state token is not in a valid state to branch off of.

#### `cosmos.state.v1.commit`

Takes no parameters and commits the current branched state token in the **message packet** against the underlying state it was branched from.

#### `cosmos.state.v1.rollback`

Takes no parameters and rolls back any changes to the branched state token in the **message packet**.

### KV Store API

#### `cosmos.kvstore.v1.get`

* Volatility: Readonly
* Input Parameter 1: `key`
* Output Parameter 1: `value`
* Errors: `key_not_found`

A 64kb packet size is suggested with the key at offset 16,384 and the value at offset 32,768.

#### `cosmos.kvstore.v1.set`

* Volatility: Volatile
* Input Parameter 1: `key`
* Input Parameter 2: `value`
* Errors: None
* Suggested packet size: 64kb

The same packet utilization as get is suggested.

#### `cosmos.kvstore.v1.delete`

* Volatility: Volatile
* Input Parameters 1: `key`
* Errors: None (should there be an error for key not found?)

The suggested packet size is 32kb with the key at offset 16,384.

#### `cosmos.kvstore.v1.has`

* Volatility: Readonly
* Input Parameter 1: `key`
* Errors: `not_found`

The same packet utilization as delete is suggested.

### Ordered KV Store API

#### `cosmos.orderedkvstore.v1.iterator`

* Volatility: Readonly
* Input Parameter 1: `start`
* Input Parameter 2: `end`
* Output Parameter 1: `iterator` - 32 bytes that are to be used as the next state token
* Errors: None

The suggested packet size is 32kb with the start key at offset 16,384 and the end key at offset 32,768,
and iterator at any offset not otherwise used.

#### `cosmos.orderedkvstore.v1.reverse_iterator`

* Volatility: Readonly
* Input Parameter 1: `start`
* Input Parameter 2: `end`
* Output Parameter 1: `iterator` - 32 bytes that are to be used as the next state token
* Errors: None

The same packet utilization as iterator is suggested.

#### `cosmos.orderedkvstore.v1.iterator_next`

* Volatility: Readonly
* Input Parameters: None (uses the state token)
* Output Parameter 1: `key`
* Output Parameter 2: `value`
* Errors: `iterator_done`

The same packet utilization as get is suggested.

#### `cosmos.orderedkvstore.v1.iterator_close`

* Volatility: Readonly
* Input Parameters: None (uses the state token)
* Errors: None

## Decision

## Consequences (optional)

### Backwards Compatibility

### Positive

### Negative

### Neutral

### References

## Discussion
