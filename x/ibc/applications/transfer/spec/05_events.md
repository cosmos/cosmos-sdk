<!--
order: 5
-->

# Events

## MsgTransfer

| Type                                       | Attribute Key | Attribute Value |
|--------------------------------------------|---------------|-----------------|
| ibc-applications-transfer-v1-EventTransfer | sender        | {sender}        |
| ibc-applications-transfer-v1-EventTransfer | receiver      | {receiver}      |
| message                                    | action        | transfer        |

## OnRecvPacket callback

| Type                                                | Attribute Key | Attribute Value |
|-----------------------------------------------------|---------------|-----------------|
| ibc-applications-transfer-v1-EventDenominationTrace | trace_hash    | {hex_hash}      |

## OnAcknowledgePacket callback (emitted any one of the below events)

| Type                                                     | Attribute Key | Attribute Value   |
|----------------------------------------------------------|---------------|-------------------|
| ibc-applications-transfer-v1-EventAcknowledgementSuccess | success       | {ack.Response}    |
| ibc-applications-transfer-v1-EventAcknowledgementError   | error         | {ack.Response}    |
