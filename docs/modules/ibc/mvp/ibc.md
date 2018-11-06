# IBC Specification

The IBC (Inter Blockchain Communication) protocol specifies how tokens, 
non-fungible assets and complex objects can be moved securely between different
zones (independent blockchains). IBC is conceptually similar to TCP/IP in the 
sense that anyone can implement it in order to be able to establish IBC
connections with willing clients.


## Terms

How IBC module treats incoming IBC packets is similar to how BaseApp treats 
incoming transactions. Therefore, the components of IBC module have their 
corresponding pair in BaseApp.

| BaseApp Terms | IBC Terms  |
| ------------- | ---------- |
| Router        | Dispatcher |
| Tx            | Packet     |
| Msg           | Payload    |


## MVP Specifications

### [MVP1](./mvp1.md)

MVP1 will contain the basic functionalities, including packet generation and 
incoming packet processing. There will be no security check for incoming 
packets.

### [MVP2](./mvp2.md)

The IBC module will be more modular in MVP2. Individual modules can register 
custom handlers on the IBC module.

### [MVP3](./mvp3.md)

Light client verification is added to verify an IBC packet from another chain. 
Registering chains with their RoT(Root of Trust) is added as well.

### [MVP4](./mvp4.md)

ACK verification / timeout handler helper functions and messaging queues are 
implemented to make it safe. Callbacks will be registered to the dispatcher to 
handle failure when they register handlers.