# Fix missing mutex protection in Info() ABCI method

## Problem

`Info()` in `baseapp/abci.go` reads shared fields (`cms`, `name`, `version`) without acquiring `app.mu`, while other methods like `InitChain()` modify these fields under the mutex. This creates a data race.

## Root Cause

The `BaseApp` struct uses `app.mu` to protect concurrent access to its fields. Methods like `InitChain()`, `FinalizeBlock()`, and `Commit()` properly lock `app.mu` before accessing shared state. However, `Info()` — which is called by CometBFT during handshake and queries — reads `app.cms`, `app.name`, and `app.version` without locking, creating a potential data race with concurrent state modifications.

## Fix

Added `app.mu.Lock()` and `defer app.mu.Unlock()` at the start of `Info()`, matching the pattern used by other ABCI methods.

## Testing

- Run the app with the Go race detector (`-race` flag).
- Trigger concurrent `Info` calls during chain initialization.
- Previously: race detector may report a data race. Now: properly synchronized.

## Impact

Affects Cosmos SDK node operators. The data race is most likely during startup (when `InitChain` and `Info` may be called concurrently) and could cause reading inconsistent app state.
