<!--
order: 6
-->

# Events

The IBC module emits the following events:

## ICS 02 - Client

### MsgCreateClient

| Type          | Attribute Key | Attribute Value |
|---------------|---------------|-----------------|
| create_client | client_id     | {clientID}      |
| message       | module        | ibc_client      |
| message       | action        | create_client   |
| message       | sender        | {signer}        |

### MsgUpdateClient

| Type          | Attribute Key | Attribute Value |
|---------------|---------------|-----------------|
| update_client | client_id     | {clientID}      |
| message       | module        | ibc_client      |
| message       | action        | update_client   |
| message       | sender        | {signer}        |

### MsgSubmitMisbehaviour

| Type                | Attribute Key | Attribute Value     |
|---------------------|---------------|---------------------|
| client_misbehaviour | client_id     | {clientID}          |
| message             | module        | evidence            |
| message             | action        | client_misbehaviour |
| message             | sender        | {signer}            |

## ICS 03 - Connection

### MsgConnectionOpenInit

| Type                 | Attribute Key          | Attribute Value         |
|----------------------|------------------------|-------------------------|
| connection_open_init | connection_id          | {connectionID}          |
| connection_open_init | client_id              | {clientID}              |
| connection_open_init | counterparty_client_id | {counterparty.clientID} |
| message              | module                 | ibc_connection          |
| message              | action                 | connection_open_init    |
| message              | sender                 | {signer}                |

### MsgConnectionOpenTry

| Type                | Attribute Key | Attribute Value     |
|---------------------|---------------|---------------------|
| connection_open_try | connection_id | {connectionID}      |
| connection_open_try | client_id     | {clientID}          |
| message             | module        | ibc_connection      |
| message             | action        | connection_open_try |
| message             | sender        | {signer}            |

### MsgConnectionOpenAck

| Type                 | Attribute Key          | Attribute Value         |
|----------------------|------------------------|-------------------------|
| connection_open_ack  | connection_id          | {connectionID}          |
| connection_open_ack  | client_id              | {clientID}              |
| connection_open_init | counterparty_client_id | {counterparty.clientID} |
| message              | module                 | ibc_connection          |
| message              | action                 | connection_open_ack     |
| message              | sender                 | {signer}                |

### MsgConnectionOpenConfirm

| Type                    | Attribute Key | Attribute Value         |
|-------------------------|---------------|-------------------------|
| connection_open_confirm | connection_id | {connectionID}          |
| message                 | module        | ibc_connection          |
| message                 | action        | connection_open_confirm |
| message                 | sender        | {signer}                |

## ICS 04 - Channel

### MsgChannelOpenInit

| Type              | Attribute Key           | Attribute Value                  |
|-------------------|-------------------------|----------------------------------|
| channel_open_init | port_id                 | {portID}                         |
| channel_open_init | channel_id              | {channelID}                      |
| channel_open_init | counterparty_port_id    | {channel.counterparty.portID}    |
| channel_open_init | counterparty_channel_id | {channel.counterparty.channelID} |
| channel_open_init | connection_id           | {channel.connectionHops}         |
| message           | module                  | ibc_channel                      |
| message           | action                  | channel_open_init                |
| message           | sender                  | {signer}                         |

### MsgChannelOpenTry

| Type             | Attribute Key           | Attribute Value                  |
|------------------|-------------------------|----------------------------------|
| channel_open_try | port_id                 | {portID}                         |
| channel_open_try | channel_id              | {channelID}                      |
| channel_open_try | counterparty_port_id    | {channel.counterparty.portID}    |
| channel_open_try | counterparty_channel_id | {channel.counterparty.channelID} |
| channel_open_try | connection_id           | {channel.connectionHops}         |
| message          | module                  | ibc_channel                      |
| message          | action                  | channel_open_try                 |
| message          | sender                  | {signer}                         |

### MsgChannelOpenAck

| Type             | Attribute Key | Attribute Value  |
|------------------|---------------|------------------|
| channel_open_ack | port_id       | {portID}         |
| channel_open_ack | channel_id    | {channelID}      |
| message          | module        | ibc_channel      |
| message          | action        | channel_open_ack |
| message          | sender        | {signer}         |

### MsgChannelOpenConfirm

| Type                 | Attribute Key | Attribute Value      |
|----------------------|---------------|----------------------|
| channel_open_confirm | port_id       | {portID}             |
| channel_open_confirm | channel_id    | {channelID}          |
| message              | module        | ibc_channel          |
| message              | action        | channel_open_confirm |
| message              | sender        | {signer}             |

### MsgChannelCloseInit

| Type               | Attribute Key | Attribute Value    |
|--------------------|---------------|--------------------|
| channel_close_init | port_id       | {portID}           |
| channel_close_init | channel_id    | {channelID}        |
| message            | module        | ibc_channel        |
| message            | action        | channel_close_init |
| message            | sender        | {signer}           |

### MsgChannelCloseConfirm

| Type                  | Attribute Key | Attribute Value       |
|-----------------------|---------------|-----------------------|
| channel_close_confirm | port_id       | {portID}              |
| channel_close_confirm | channel_id    | {channelID}           |
| message               | module        | ibc_channel           |
| message               | action        | channel_close_confirm |
| message               | sender        | {signer}              |
