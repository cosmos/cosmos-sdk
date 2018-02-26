## 5 Conclusion

We have demonstrated a secure, performant, and flexible protocol for connecting 
two blockchains with complete finality using a secure, reliable messaging 
queue. The algorithm and semantics of all data types have been defined above, 
which provides a solid basis for reasoning about correctness and efficiency of 
the algorithm.

The observant reader may note that while we have defined a message queue 
protocol, we have not yet defined how to use that to transfer value within the 
Cosmos ecosystem. We will shortly release a separate paper on Cosmos IBC that 
defines the application logic used for direct value transfer as well as routing
over the Cosmos hub. That paper builds upon the IBC protocol defined here and 
provides a first example of how to reason about application logic and global 
invariants in the context of IBC.

There is a reference implementation of the Cosmos IBC protocol as part of the 
Cosmos SDK, written in go and freely usable under the Apache license. For those 
wish to write an implementation of IBC in another language, or who want to 
analyze the specification further, the following appendixes define the exact 
message formats and binary encoding.
