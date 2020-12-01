<!--
order: 6
-->

# Events

The IBC module emits the following events. It can be expected that the type `message`,
with an attirbute key of `action` will represent the first event for each message
being processed as emitted by the SDK's baseapp.

All the events for the Channel handshakes, `SendPacket`, `RecvPacket`, `AcknowledgePacket`, 
`TimeoutPacket` and `TimeoutOnClose` will emit additional events not specified here due to
callbacks to IBC applications.

## ICS 02 - Client

### MsgCreateClient

| Type                                 | Attribute Key    | Attribute Value   |
|--------------------------------------|------------------|-------------------|
| ibc-core-client-v1-EventCreateClient | client_id        | {clientId}        |
| ibc-core-client-v1-EventCreateClient | client_type      | {clientType}      |
| ibc-core-client-v1-EventCreateClient | consensus_height | {consensusHeight} |
| message                              | action           | create_client     |

### MsgUpdateClient

| Type                                 | Attribute Key    | Attribute Value   |
|--------------------------------------|------------------|-------------------|
| ibc-core-client-v1-EventUpdateClient | client_id        | {clientId}        |
| ibc-core-client-v1-EventUpdateClient | client_type      | {clientType}      |
| ibc-core-client-v1-EventUpdateClient | consensus_height | {consensusHeight} |
| message                              | action           | update_client     |

### MsgUpgradeClient

| Type                                  | Attribute Key    | Attribute Value   |
|---------------------------------------|------------------|-------------------|
| ibc-core-client-v1-EventUpgradeClient | client_id        | {clientId}        |
| ibc-core-client-v1-EventUpgradeClient | client_type      | {clientType}      |
| ibc-core-client-v1-EventUpgradeClient | consensus_height | {consensusHeight} |
| message                               | action           | upgrade_client    |

### MsgSubmitMisbehaviour

| Type                                       | Attribute Key    | Attribute Value     |
|--------------------------------------------|------------------|---------------------|
| ibc-core-client-v1-EventClientMisbehaviour | client_id        | {clientId}          |
| ibc-core-client-v1-EventClientMisbehaviour | client_type      | {clientType}        |
| ibc-core-client-v1-EventClientMisbehaviour | consensus_height | {consensusHeight}   |
| message                                    | action           | client_misbehaviour |
| message                                    | module           | evidence            |
| message                                    | sender           | {senderAddress}     |
| submit_evidence                            | evidence_hash    | {evidenceHash}      |

### UpdateClientProposal

| Type                                         | Attribute Key    | Attribute Value   |
|----------------------------------------------|------------------|-------------------|
| ibc-core-client-v1-EventUpdateClientProposal | client_id        | {clientId}        |
| ibc-core-client-v1-EventUpdateClientProposal | client_type      | {clientType}      |
| ibc-core-client-v1-EventUpdateClientProposal | consensus_height | {consensusHeight} |



## ICS 03 - Connection

### MsgConnectionOpenInit

| Type                                           | Attribute Key              | Attribute Value             |
|------------------------------------------------|----------------------------|-----------------------------|
| ibc-core-connection-v1-EventConnectionOpenInit | connection_id              | {connectionId}              |
| ibc-core-connection-v1-EventConnectionOpenInit | client_id                  | {clientId}                  |
| ibc-core-connection-v1-EventConnectionOpenInit | counterparty_client_id     | {counterparty.clientId}     |
| message                                        | action                     | connection_open_init        |

### MsgConnectionOpenTry

| Type                                          | Attribute Key              | Attribute Value             |
|-----------------------------------------------|----------------------------|-----------------------------|
| ibc-core-connection-v1-EventConnectionOpenTry | connection_id              | {connectionId}              | 
| ibc-core-connection-v1-EventConnectionOpenTry | client_id                  | {clientId}                  |
| ibc-core-connection-v1-EventConnectionOpenTry | counterparty_client_id     | {counterparty.clientId      |
| ibc-core-connection-v1-EventConnectionOpenTry | counterparty_connection_id | {counterparty.connectionId} |
| message                                       | action                     | connection_open_try         |

### MsgConnectionOpenAck

| Type                                           | Attribute Key              | Attribute Value             |
|------------------------------------------------|----------------------------|-----------------------------|
| ibc-core-connection-v1-EventConnectionOpenAck  | connection_id              | {connectionId}              |
| ibc-core-connection-v1-EventConnectionOpenAck  | client_id                  | {clientId}                  |
| ibc-core-connection-v1-EventConnectionOpenAck  | counterparty_client_id     | {counterparty.clientId}     |
| ibc-core-connection-v1-EventConnectionOpenAck  | counterparty_connection_id | {counterparty.connectionId} |
| message                                        | action                     | connection_open_ack         |

### MsgConnectionOpenConfirm

| Type                                              | Attribute Key              | Attribute Value             |
|---------------------------------------------------|----------------------------|-----------------------------|
| ibc-core-connection-v1-EventConnectionOpenConfirm | connection_id              | {connectionId}              |
| ibc-core-connection-v1-EventConnectionOpenConfirm | client_id                  | {clientId}                  |
| ibc-core-connection-v1-EventConnectionOpenConfirm | counterparty_client_id     | {counterparty.clientId}     |
| ibc-core-connection-v1-EventConnectionOpenConfirm | counterparty_connection_id | {counterparty.connectionId} |
| message                                           | action                     | connection_open_confirm     |

## ICS 04 - Channel

### MsgChannelOpenInit

| Type                                     | Attribute Key           | Attribute Value                  |
|------------------------------------------|-------------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelOpenInit | port_id                 | {portId}                         |
| ibc-core-channel-v1-EventChannelOpenInit | channel_id              | {channelId}                      |
| ibc-core-channel-v1-EventChannelOpenInit | counterparty_port_id    | {channel.counterparty.portId}    |
| ibc-core-channel-v1-EventChannelOpenInit | connection_id           | {channel.connectionHops}         |
| message                                  | action                  | channel_open_init                |

### MsgChannelOpenTry

| Type                                    | Attribute Key           | Attribute Value                  |
|-----------------------------------------|-------------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelOpenTry | port_id                 | {portId}                         |
| ibc-core-channel-v1-EventChannelOpenTry | channel_id              | {channelId}                      |
| ibc-core-channel-v1-EventChannelOpenTry | counterparty_port_id    | {channel.counterparty.portId}    |
| ibc-core-channel-v1-EventChannelOpenTry | counterparty_channel_id | {channel.counterparty.channelId} |
| ibc-core-channel-v1-EventChannelOpenTry | connection_id           | {channel.connectionHops}         |
| message                                 | action                  | channel_open_try                 |

### MsgChannelOpenAck

| Type                                    | Attribute Key           | Attribute Value                  |
|-----------------------------------------|-------------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelOpenAck | port_id                 | {portId}                         |
| ibc-core-channel-v1-EventChannelOpenAck | channel_id              | {channelId}                      |
| ibc-core-channel-v1-EventChannelOpenAck | counterparty_port_id    | {channel.counterparty.portId}    |
| ibc-core-channel-v1-EventChannelOpenAck | counterparty_channel_id | {channel.counterparty.channelId} |
| ibc-core-channel-v1-EventChannelOpenAck | connection_id           | {channel.connectionHops}         |
| message                                 | action                  | channel_open_ack                 |

### MsgChannelOpenConfirm

| Type                                        | Attribute Key           | Attribute Value                  |
|---------------------------------------------|-------------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelOpenConfirm | port_id                 | {portId}                         |
| ibc-core-channel-v1-EventChannelOpenConfirm | channel_id              | {channelId}                      |
| ibc-core-channel-v1-EventChannelOpenConfirm | counterparty_port_id    | {channel.counterparty.portId}    |
| ibc-core-channel-v1-EventChannelOpenConfirm | counterparty_channel_id | {channel.counterparty.channelId} |
| ibc-core-channel-v1-EventChannelOpenConfirm | connection_id           | {channel.connectionHops}         |
| message                                     | action                  | channel_open_confirm             |

### MsgChannelCloseInit

| Type                                      | Attribute Key           | Attribute Value                  |
|-------------------------------------------|-------------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelCloseInit | port_id                 | {portId}                         |
| ibc-core-channel-v1-EventChannelCloseInit | channel_id              | {channelId}                      |
| ibc-core-channel-v1-EventChannelCloseInit | counterparty_port_id    | {channel.counterparty.portId}    |
| ibc-core-channel-v1-EventChannelCloseInit | counterparty_channel_id | {channel.counterparty.channelId} |
| ibc-core-channel-v1-EventChannelCloseInit | connection_id           | {channel.connectionHops}         |
| message                                   | action                  | channel_close_init               |

### MsgChannelCloseConfirm

| Type                                         | Attribute Key           | Attribute Value                  |
|----------------------------------------------|-------------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelCloseConfirm | port_id                 | {portId}                         |
| ibc-core-channel-v1-EventChannelCloseConfirm | channel_id              | {channelId}                      |
| ibc-core-channel-v1-EventChannelCloseConfirm | counterparty_port_id    | {channel.counterparty.portId}    |
| ibc-core-channel-v1-EventChannelCloseConfirm | counterparty_channel_id | {channel.counterparty.channelId} |
| ibc-core-channel-v1-EventChannelCloseConfirm | connection_id           | {channel.connectionHops}         |
| message                                      | action                  | channel_close_confirm            |

### SendPacket (application module call)

| Type                                       | Attribute Key     | Attribute Value                  |
|--------------------------------------------|-------------------|----------------------------------|
| ibc-core-channel-v1-EventChannelSendPacket | data              | {data}                           |
| ibc-core-channel-v1-EventChannelSendPacket | timeout_height    | {timeoutHeight}                  |
| ibc-core-channel-v1-EventChannelSendPacket | timeout_timestamp | {timeoutTimestamp}               |
| ibc-core-channel-v1-EventChannelSendPacket | sequence          | {sequence}                       |
| ibc-core-channel-v1-EventChannelSendPacket | src_port          | {sourcePort}                     |
| ibc-core-channel-v1-EventChannelSendPacket | src_channel       | {sourceChannel}                  |
| ibc-core-channel-v1-EventChannelSendPacket | dst_port          | {destinationPort}                |
| ibc-core-channel-v1-EventChannelSendPacket | dst_channel       | {destinationChannel}             |
| ibc-core-channel-v1-EventChannelSendPacket | channel_ordering  | {channel.Ordering}               |
| message                                    | action            | application-module-defined-field |

### MsgRecvPacket 

| Type                                       | Attribute Key     | Attribute Value      |
|--------------------------------------------|-------------------|----------------------|
| ibc-core-channel-v1-EventChannelRecvPacket | data              | {data}               |
| ibc-core-channel-v1-EventChannelRecvPacket | timeout_height    | {timeoutHeight}      |
| ibc-core-channel-v1-EventChannelRecvPacket | timeout_timestamp | {timeoutTimestamp}   |
| ibc-core-channel-v1-EventChannelRecvPacket | sequence          | {sequence}           |
| ibc-core-channel-v1-EventChannelRecvPacket | src_port          | {sourcePort}         |
| ibc-core-channel-v1-EventChannelRecvPacket | src_channel       | {sourceChannel}      |
| ibc-core-channel-v1-EventChannelRecvPacket | dst_port          | {destinationPort}    |
| ibc-core-channel-v1-EventChannelRecvPacket | dst_channel       | {destinationChannel} |
| ibc-core-channel-v1-EventChannelRecvPacket | channel_ordering  | {channel.Ordering}   |
| message                                    | action            | recv_packet          |

### MsgAcknowledgePacket 

| Type                                      | Attribute Key     | Attribute Value      |
|-------------------------------------------|-------------------|----------------------|
| ibc-core-channel-v1-EventChannelAckPacket | data              | {data}               |
| ibc-core-channel-v1-EventChannelAckPacket | timeout_height    | {timeoutHeight}      |
| ibc-core-channel-v1-EventChannelAckPacket | timeout_timestamp | {timeoutTimestamp}   |
| ibc-core-channel-v1-EventChannelAckPacket | sequence          | {sequence}           |
| ibc-core-channel-v1-EventChannelAckPacket | src_port          | {sourcePort}         |
| ibc-core-channel-v1-EventChannelAckPacket | src_channel       | {sourceChannel}      |
| ibc-core-channel-v1-EventChannelAckPacket | dst_port          | {destinationPort}    |
| ibc-core-channel-v1-EventChannelAckPacket | dst_channel       | {destinationChannel} |
| ibc-core-channel-v1-EventChannelAckPacket | channel_ordering  | {channel.Ordering}   |
| ibc-core-channel-v1-EventChannelAckPacket | acknowledgement   | {acknowledgement}    |
| message                                   | action            | acknowledge_packet   |

### MsgTimeoutPacket & MsgTimeoutOnClose 

| Type                                          | Attribute Key     | Attribute Value      |
|-----------------------------------------------|-------------------|----------------------|
| ibc-core-channel-v1-EventChannelTimeoutPacket | data              | {data}               |
| ibc-core-channel-v1-EventChannelTimeoutPacket | timeout_height    | {timeoutHeight}      |
| ibc-core-channel-v1-EventChannelTimeoutPacket | timeout_timestamp | {timeoutTimestamp}   |
| ibc-core-channel-v1-EventChannelTimeoutPacket | sequence          | {sequence}           |
| ibc-core-channel-v1-EventChannelTimeoutPacket | src_port          | {sourcePort}         |
| ibc-core-channel-v1-EventChannelTimeoutPacket | src_channel       | {sourceChannel}      |
| ibc-core-channel-v1-EventChannelTimeoutPacket | dst_port          | {destinationPort}    |
| ibc-core-channel-v1-EventChannelTimeoutPacket | dst_channel       | {destinationChannel} |
| ibc-core-channel-v1-EventChannelTimeoutPacket | channel_ordering  | {channel.Ordering}   |
| message                                       | action            | timeout_packet       |

### WriteAcknowledgement 

| Type                                     | Attribute Key     | Attribute Value      |
|------------------------------------------|-------------------|----------------------|
| ibc-core-channel-v1-EventChannelWriteAck | data              | {data}               |
| ibc-core-channel-v1-EventChannelWriteAck | timeout_height    | {timeoutHeight}      |
| ibc-core-channel-v1-EventChannelWriteAck | timeout_timestamp | {timeoutTimestamp}   |
| ibc-core-channel-v1-EventChannelWriteAck | sequence          | {sequence}           |
| ibc-core-channel-v1-EventChannelWriteAck | src_port          | {sourcePort}         |
| ibc-core-channel-v1-EventChannelWriteAck | src_channel       | {sourceChannel}      |
| ibc-core-channel-v1-EventChannelWriteAck | dst_port          | {destinationPort}    |
| ibc-core-channel-v1-EventChannelWriteAck | dst_channel       | {destinationChannel} |
| ibc-core-channel-v1-EventChannelWriteAck | acknowledgement   | {acknowledgement}    |


