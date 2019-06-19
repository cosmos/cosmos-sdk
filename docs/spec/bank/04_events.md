# Events

The bank module emits the following events:

## Handlers

### MsgSend

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| transfer | sender        | {senderAddress}    |
| transfer | recipient     | {recipientAddress} |
| message  | module        | bank               |
| message  | action        | send               |

### MsgMultiSend

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| transfer | sender        | {senderAddress}    |
| transfer | recipient     | {recipientAddress} |
| message  | module        | bank               |
| message  | action        | multisend          |
