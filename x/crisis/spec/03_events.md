<!--
order: 3
-->

# Events

The crisis module emits the following events:

## MsgServer

### MsgVerifyInvariant

| Type      | Attribute Key | Attribute Value  |
|-----------|---------------|------------------|
| invariant | route         | {invariantRoute} |
| message   | module        | crisis           |
| message   | action        | verify_invariant |
| message   | sender        | {senderAddress}  |
