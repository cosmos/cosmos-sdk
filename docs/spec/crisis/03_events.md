# Events

The crisis module emits the following events:

## Handlers

### MsgVerifyInvariance

| Type      | Attribute Key | Attribute Value  |
|-----------|---------------|------------------|
| invariant | route         | {invariantRoute} |
| invariant | sender        | {senderAddress}  |
| message   | module        | crisis           |
| message   | action        | verify_invariant |
