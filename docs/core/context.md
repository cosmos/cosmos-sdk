# Context

## Prerequisites

* [Anatomy of an SDK Application](../basics/app-anatomy.md)
* [Lifecycle of a Transaction](../basics/tx-lifecycle.md)

## Synopsis

This document details the SDK `context` type.

- [Go Context Package](#go-context-package)
- [SDK Context Type Definition](#context-definition)
- [Cache Wrapping](#cache-wrapping)
- [Usage](#usage)

## Go Context Package

## Context Definition

```go
type Context struct {
	ctx           context.Context
	ms            MultiStore
	header        abci.Header
	chainID       string
	txBytes       []byte
	logger        log.Logger
	voteInfo      []abci.VoteInfo
	gasMeter      GasMeter
	blockGasMeter GasMeter
	checkTx       bool
	minGasPrice   DecCoins
	consParams    *abci.ConsensusParams
	eventManager  *EventManager
}
```

- **Context**
- **Multistore**
- **ABCI Header**
- **Chain ID**
- **Transaction Bytes**
- **Logger**
- **ABCI Vote Info**
- **Gas Meters**
- **CheckTx Mode**
- **Node Min Gas Price**
- **Consensus Params**
- **Event Manager**

## Cache Wrapping

## Usage

Pattern: receive a context, cache wrap (make a copy), make changes to copy, commit if successful, return.
