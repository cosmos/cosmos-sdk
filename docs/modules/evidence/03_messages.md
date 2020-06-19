<!--
order: 3
-->

# Messages

## MsgSubmitEvidence

Evidence is submitted through a `MsgSubmitEvidence` message:

```go
type MsgSubmitEvidence struct {
  Evidence  Evidence
  Submitter AccAddress
}
```

Note, the `Evidence` of a `MsgSubmitEvidence` message must have a corresponding
`Handler` registered with the `x/evidence` module's `Router` in order to be processed
and routed correctly.

Given the `Evidence` is registered with a corresponding `Handler`, it is processed
as follows:

```go
func SubmitEvidence(ctx Context, evidence Evidence) error {
  if _, ok := GetEvidence(ctx, evidence.Hash()); ok {
    return ErrEvidenceExists(codespace, evidence.Hash().String())
  }
  if !router.HasRoute(evidence.Route()) {
    return ErrNoEvidenceHandlerExists(codespace, evidence.Route())
  }

  handler := router.GetRoute(evidence.Route())
  if err := handler(ctx, evidence); err != nil {
    return ErrInvalidEvidence(codespace, err.Error())
  }

  SetEvidence(ctx, evidence)
  return nil
}
```

First, there must not already exist valid submitted `Evidence` of the exact same
type. Secondly, the `Evidence` is routed to the `Handler` and executed. Finally,
if there is no error in handling the `Evidence`, it is persisted to state.
