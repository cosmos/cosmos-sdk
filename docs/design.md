## Design philosphy

The design of the Cosmos SDK is based on the principles of "cababilities systems".

TODO If you see this on the sdk2 branch, it's because I'm still expanding this high-level section.

Sections:

* Introduction
 - Note to skip to Basecoin example to dive into code.
* Capabilities systems
 - http://www.erights.org/elib/capability/ode/ode.pdf
* Tx & Msg
* MultiStore
* Context & Handler
* AnteHandler
 - Fees
 - Authentication
* Accounts and x/auth
* AccountStore
* Wire codec
 - vs encoding/json
 - vs protobuf
* Dummy example
* Basecoin example
* Conclusion

## ######################################
## TODO bring the below up.

More details about the design goals of particular components follows.

### Store

- Fast balanced dynamic Merkle tree for storing application state
- Support multiple Merkle tree backends in a single store 
    - allows using Ethereum Patricia Trie and Tendermint IAVL in same app
- Support iteration
- Provide caching for intermediate state during execution of blocks and transactions (including for iteration)
- Retain some amount of recent historical state
- Allow many kinds of proofs (exists, absent, range, etc.) on current and retained historical state

### ABCI Application

- Simple connector between developer's application logic and the ABCI protocol
- Simplify discrepancy between DeliverTx and CheckTx
- Handles ABCI handshake logic and historical state
- Provide simple hooks to BeginBlock and EndBlock 

### Transaction Processing

- Transactions consist of composeable messages 
- Processing via series of handlers that handle authenticate, deduct fees, transfer coins, etc.
- Developers control tx encoding
    - Default go-wire
    - Must be able to write eg. Ethermint using the SDK with Ethereum-native transaction encoding
- Handler access to the store is restricted via capabilities and interfaces
- Context object holds

### Data Types

- Default Ethereum-style Account 
- Default multi-asset Coins



