## 5 Conclusion

([Back to table of contents](README.md#contents))

We have demonstrated a secure, performant, and flexible protocol for cross-blockchain messaging, and provided sufficient detail to reason about the correctness and efficiency of the protocol.

This document defines solely a message queue protocol - not the application-level semantics which must sit on top of it to enable asset transfer between two chains. We will shortly release a separate paper on Cosmos IBC that defines the application logic used for direct value transfer as well as routing over the Cosmos hub. That paper builds upon the IBC protocol defined here and provides a first example of how to reason about application logic and global invariants in the context of IBC.

There is a reference implementation of the Cosmos IBC protocol as part of the Cosmos SDK, written in Golang and released under the Apache license. To facilitate implementations in other langauages which are amino-compatible with the Cosmos implementation, the following appendices define exact message and binary encoding formats.
