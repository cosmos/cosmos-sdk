# AppManager Documentation

The AppManager serves as a high-level coordinator, delegating most operations to the STF while managing state access through the Store interface.

This document outlines the main external calls in the AppManager package, their execution flows, and dependencies.

## Table of Contents
- [InitGenesis](#initgenesis)
- [ExportGenesis](#exportgenesis)
- [DeliverBlock](#deliverblock)
- [ValidateTx](#validatetx)
- [Simulate](#simulate)
- [SimulateWithState](#simulatewithstate)
- [Query](#query)
- [QueryWithState](#querywithstate)

## InitGenesis

InitGenesis initializes the genesis state of the application.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant InitGenesisImpl
participant STF
participant State
Caller->>AppManager: InitGenesis(ctx, blockRequest, genesisJSON, decoder)
AppManager->>InitGenesisImpl: initGenesis(ctx, genesisJSON, txHandler)
loop For each genesis transaction
InitGenesisImpl->>InitGenesisImpl: Decode and collect transactions
end
InitGenesisImpl-->>AppManager: genesisState, validatorUpdates, error
AppManager->>STF: DeliverBlock(ctx, blockRequest, genesisState)
STF-->>AppManager: blockResponse, blockZeroState, error
AppManager->>State: Apply state changes
AppManager-->>Caller: blockResponse, genesisState, error
```

### Dependencies
- Required Input:
  - Context
  - BlockRequest
  - Genesis JSON
  - Transaction decoder
- Required Components:
  - InitGenesis implementation
  - STF
  - Store interface

## ExportGenesis

ExportGenesis exports the current application state as genesis state.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant ExportGenesisImpl
Caller->>AppManager: ExportGenesis(ctx, version)
AppManager->>ExportGenesisImpl: exportGenesis(ctx, version)
ExportGenesisImpl-->>Caller: genesisJSON, error
```

### Dependencies
- Required Input:
  - Context
  - Version
- Required Components:
  - ExportGenesis implementation
  - Store interface


## DeliverBlock

DeliverBlock processes a block of transactions.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant Store
participant STF
Caller->>AppManager: DeliverBlock(ctx, block)
AppManager->>Store: StateLatest()
Store-->>AppManager: version, currentState, error
AppManager->>STF: DeliverBlock(ctx, block, currentState)
STF-->>Caller: blockResponse, newState, error
```


### Dependencies
- Required Input:
  - Context
  - BlockRequest
- Required Components:
  - Store interface
  - STF

## ValidateTx

ValidateTx validates a transaction against the latest state.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant Store
participant STF
Caller->>AppManager: ValidateTx(ctx, tx)
AppManager->>Store: StateLatest()
Store-->>AppManager: version, latestState, error
AppManager->>STF: ValidateTx(ctx, latestState, gasLimit, tx)
STF-->>Caller: TxResult, error
```


### Dependencies
- Required Input:
  - Context
  - Transaction
- Required Components:
  - Store interface
  - STF
  - Configuration (for gas limits)

## Simulate

Simulate executes a transaction simulation using the latest state.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant Store
participant STF
Caller->>AppManager: Simulate(ctx, tx)
AppManager->>Store: StateLatest()
Store-->>AppManager: version, state, error
AppManager->>STF: Simulate(ctx, state, gasLimit, tx)
STF-->>Caller: TxResult, WriterMap, error
```

### Dependencies
- Required Input:
  - Context
  - Transaction
- Required Components:
  - Store interface
  - STF
  - Configuration (for gas limits)  

## SimulateWithState

SimulateWithState executes a transaction simulation using provided state.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant STF
Caller->>AppManager: SimulateWithState(ctx, state, tx)
AppManager->>STF: Simulate(ctx, state, gasLimit, tx)
STF-->>Caller: TxResult, WriterMap, error
```

### Dependencies
- Required Input:
  - Context
  - Transaction
  - State

  ## Query

Query executes a query at a specific version.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant Store
participant STF
Caller->>AppManager: Query(ctx, version, request)
alt version == 0
AppManager->>Store: StateLatest()
else version > 0
AppManager->>Store: StateAt(version)
end
Store-->>AppManager: queryState, error
AppManager->>STF: Query(ctx, queryState, gasLimit, request)
STF-->>Caller: response, error
```

### Dependencies
- Required Input:
  - Context
  - Version (or 0 for latest)
  - Query request
- Required Components:
  - Store interface
  - STF
  - Configuration (for gas limits)

  ## QueryWithState

QueryWithState executes a query using provided state.

```mermaid
sequenceDiagram
participant Caller
participant AppManager
participant STF
Caller->>AppManager: QueryWithState(ctx, state, request)
AppManager->>STF: Query(ctx, state, gasLimit, request)
STF-->>Caller: response, error
```

### Dependencies
- Required Input:
  - Context
  - ReaderMap state
  - Query request
- Required Components:
  - STF
  - Configuration (for gas limits)

## Common Dependencies

All operations depend on:
- Context management
- Error handling
- Gas metering
- State management (Store interface)
- STF interface