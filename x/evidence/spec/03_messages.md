<!--
order: 3
-->

# Messages

## MsgSubmitEvidence

Evidence is submitted through a `MsgSubmitEvidence` message:

```protobuf
// MsgSubmitEvidence represents a message that supports submitting arbitrary
// Evidence of misbehavior such as equivocation or counterfactual signing.
message MsgSubmitEvidence {
  string              submitter = 1;
  google.protobuf.Any evidence  = 2;
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
    return sdkerrors.Wrap(types.ErrEvidenceExists, evidence.Hash().String())
  }
  if !router.HasRoute(evidence.Route()) {
    return sdkerrors.Wrap(types.ErrNoEvidenceHandlerExists, evidence.Route())
  }

  handler := router.GetRoute(evidence.Route())
  if err := handler(ctx, evidence); err != nil {
    return sdkerrors.Wrap(types.ErrInvalidEvidence, err.Error())
  }

  ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitEvidence,
			sdk.NewAttribute(types.AttributeKeyEvidenceHash, evidence.Hash().String()),
		),
	)

  SetEvidence(ctx, evidence)
  return nil
}
```

First, there must not already exist valid submitted `Evidence` of the exact same
type. Secondly, the `Evidence` is routed to the `Handler` and executed. Finally,
if there is no error in handling the `Evidence`, an event is emitted and it is persisted to state.
