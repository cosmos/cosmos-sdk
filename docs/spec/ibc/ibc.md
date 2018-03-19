# IBC Specification

IBC(Inter-Blockchain Communication) protocol is used by multiple zones on Cosmos. Using IBC, the zones can send coins or arbitrary data to other zones.

## Terms

How IBC module treats incoming IBC packets is simillar with how BaseApp treats incoming transactions. Therefore, the components of IBC module have their corresponding pair in BaseApp.

| BaseApp Terms | IBC Terms  |
| ------------- | ---------- |
| Router        | Dispatcher |
| Tx            | Packet     |
| Msg           | Payload    |

## MVP Specifications

### [MVP1](./mvp1.md)

MVP1 will contain the basic functionalities, including packet generation and packet receivement. There will be no security check for incoming packets.

### [MVP2](./mvp2.md)

IBC module will be more modular in MVP2. Indivisual modules can register custom handlers to IBC module.

### [MVP3](./mvp3.md)

Light client verification is added to verify the message from the other chain. Registering chains with their ROT(Root Of Trust) is needed.

### [MVP4](./mvp4.md)

ACK verification / timeout handler helper functions and messaging queue are implemented to make it failsafe. Callbacks will be registered to the dispatcher to handle failure when they register handlers.
