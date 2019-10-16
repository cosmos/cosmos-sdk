# Events

The bank module emits the following events:

## Handlers

### MsgSend

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| transfer | recipient     | {recipientAddress} |
| transfer | amount        | {amount}           |
| message  | module        | bank               |
| message  | action        | send               |
| message  | sender        | {senderAddress}    |

### MsgMultiSend

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| transfer | recipient     | {recipientAddress} |
| message  | module        | bank               |
| message  | action        | multisend          |
| message  | sender        | {senderAddress}    |
