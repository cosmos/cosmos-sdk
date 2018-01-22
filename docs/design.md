## Design philosphy

The design of the Cosmos SDK is based on the principles of "cababilities systems".
TODO If you see this on the sdk2 branch, it's because I'm still expanding this high-level section.

Sections:

* Introduction
  - Note to skip to Basecoin example to dive into code.
* Capabilities systems
  - Need for module isolation
  - Capability is implied permission
  - http://www.erights.org/elib/capability/ode/ode.pdf
* Tx & Msg
* MultiStore
  - MultiStore is like a filesystem
  - Mounting an IAVLStore
* Context & Handler
* AnteHandler
  - Handling Fee payment
  - Handling Authentication
* Accounts and x/auth
  - sdk.Account
  - auth.BaseAccount
  - auth.AccountMapper
* Wire codec
  - vs encoding/json
  - vs protobuf
* Dummy example
* Basecoin example
* Conclusion
