# Tags

The bank module emits the following events/tags:

## Handlers

### MsgSend

| Key         | Value                     |
|-------------|---------------------------|
| `action`    | `send`                    |
| `category`  | `bank`                    |
| `sender`    | {senderAccountAddress}    |
| `recipient` | {recipientAccountAddress} |

### MsgMultiSend

| Key         | Value                     |
|-------------|---------------------------|
| `action`    | `multisend`               |
| `category`  | `bank`                    |
| `sender`    | {senderAccountAddress}    |
| `recipient` | {recipientAccountAddress} |
