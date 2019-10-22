# ADR 015: IBC Packet Receiver

## Changelog

- 2019 Oct 22: Initial Draft

## Context

[ICS 26](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module) defines function `handlePacketRecv`. 
`handlePacketRecv` executes per-module `onRecvPacket`, verifies the packet merkle proof, and pushes the acknowledgement bytes
to the state. 

The mechanism is simillar with the transaction handling logic in `baseapp`. After authentication, the handler is executed, and 
the authentication state change must be committed regardless of the result of the handler execution. 

ICS 26 also requires acknowledgement writing which has to be done after the handler execution and also must be commited 
regardless of the result of the handler execution.  

## Decision

We will define an `AnteHandler` for IBC packet receiving. The `AnteHandler` will iterate over the messages included in the 
transaction, type asserts to check whether the message is type of IBC packet receiving, and verifies merkle proof. 

```go
// Pseudocode
func AnteHandler(ctx sdk.Context, tx sdk.Tx, bool) (sdk.Context, sdk.Result, bool) {
  for _, msg := range tx.GetMsgs() {
    if msg, ok := msg.(MsgPacket); ok {
      if err := VerifyPacket(msg.Packet, msg.Proofs, msg.ProofHeight); err != nil {
         return ctx, err.Result(), true
      }
    }
  }
  return ctx, sdk.Result{}, false
}
```

The `AnteHandler` will be inserted to the top level application, next to the signature authentication logic provided by `auth.NewAnteHandler`.

We will define a new type of handler, named `FoldHandler`. `FoldHandler` will be called after all of the messages are executed. 

```go
type FoldHandler func(ctx sdk.Context, tx sdk.Tx, result sdk.Result) (sdk.Result, bool)

// Pseudocode
func (app *BaseApp) runTx(tx sdk.Tx) sdk.Result {
  anteCtx, msCache := app.cacheTxContext(ctx, tx)
  newCtx, result, abort := app.AnteHandler(ctx, tx)
  if abort {
    return result
  }
  msCache.Write()
  
  runMsgCtx, msCache := app.cacheTxContext(ctx, tx)
  result = app.runMsgs(runMsgCtx, msgs)
  if result.IsOk() {
    msCache.Write()
  }
  
  // BEGIN modification made in this proposal
  
  foldCtx, msCache := app.cacheTxContext(ctx, tx)
  result, abort = app.FoldHandler(ctx, tx, result)
  if abort {
     return result
  }
  msCache.Write()
  return result
  
  // END modification made in this proposal
}
```

The IBC module will expose `FoldHandler` as defined below:

```go
// Pseudocode
func FoldHandler(ctx sdk.Context, tx sdk.Tx, result sdk.Result) (sdk.Result, bool) {
  // TODO: consider multimsg
}
```

## Status

Proposed

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## References

- {reference link}
