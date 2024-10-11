# Major Design Decisions

## Account IDs vs Addresses

Originally we were planning to use an `Address` type with 20-32 variable bytes,
but this varies between VMs, in particular the EVM and Cosmos.
Now we are assigning integer account IDs and considering addresses as a pointer
to an account ID which exists at the transaction level.
There could be multiple addresses pointing to the same account ID.

**Outstanding Issues**
* What is the correct size for account IDs? Is 64 bits enough or should we use 128 bits? Ideally, we want to be able to assign account IDs concurrently without needing to lock around an incremental account ID sequence. This will require a different assignment mechanism, possibly we can take 32-bits of the transaction hash and then have a 32-bit sequence number scoped to the 32-bit transaction hash.

## Message Selectors & Type IDs

When referencing a message we can either specify the full message name or use some bytes that represent the message name in a compressed way, similar to how Ethereum uses 32-bit function selectors which are based on the hash of the function signature.
This creates more efficient message packets
and is more efficient to route to the correct handler.
The downside of a 32-bit function selector is that there can be collisions.
We can improve the collision resistance by using a 64-bit message selector.

This scheme could be extended to represent type IDs for event structs and other types,
similar to the original amino encoding (which used 4 or 7 bytes for type IDs).

64-bits is probably sufficient for uniqueness, although there could theoretically be collisions if we had to resolve such IDs globally rather than scoped to a specific module. This could maybe, someday be an issue if we have first-class modules.
However, the current implementation doesn't have first-class modules, and we may try to avoid them, so this isn't currently an issue.

## Resolving Accounts IDs

Let's define a first-class module as a handler which:
1. can only be instantiated once (singleton)
2. is the only handler for a given set of message names (i.e. no two modules can handle the same message)

In the current design, we don't have first-class modules so if we want to call `MsgSend` as in the current SDK,
we would need to know the account ID of "bank" before we can send the message. (Whether we can have singleton
handlers is a separate, but related question.)

To resolve account IDs, here are some options:
1. **Hard Code Reserved Range:** hard code account IDs for module-like things in some "reserved" ID range (ex. 1-65535) - this has the downside that it wouldn't be portable between chains
2. **Build-time Config Files:** use config files to map account aliases (ex. "bank") to account IDs - this could be portable if we find a way to configure this with different IDs at build time without needing to fork the code
3. **Runtime Config Files:** bind account aliases using a config map at runtime. This would require bundling a config descriptor with the compiled code rather than building it into the binary
4. **Runtime Module Name Resolver:** bind account aliases at runtime using some on-chain list of module names. This has the downside of binding specific APIs to specific module names.
5. **Runtime Message/Service Name Resolver:** bind account aliases at runtime using some on-chain list of service or message names (ex. "cosmos.bank.v1" or "cosmos.bank.v1.MsgSend"). This is almost the way we do it in the current SDK except we'd be resolving a default account ID rather than not knowing the account ID at all.
6. **First-class Module Messages:** meaning we don't need to know any account number, we just need to know the message name, and it will get routed to resolved account. This is different from 5 in that we don't resolve the account ID at all.

It's worth noting that options 1-5 vs any kind of first-class notion of modules have the following advantages:
* message selectors do not need to be globally unique, just unique to an account. This could be used in a rare cases to disambiguate a collision
* the resolved account ID can be used to authenticate hook callbacks. i.e. if I want to implement a bank `OnReceive` hook, I can authenticate
that the real "bank" is the caller of the hook because I know its account ID.

## Handler IDs

The basic model we are working with is that there is a root mapping of `account ID -> handler ID`
where handler ID is a pointer to the actual code to run the account in some virtual machine.

A few issues to consider:
* handler IDs should be part of consensus, but different nodes may actually use slightly different
implementations of the same handler or VM.
For instance, it shouldn't be an error to run a functionally identical handler written in Rust or Go,
built into the binary or dynamically linked, as long as there is otherwise consensus that it is the same.
It should also be possible to use different Wasm or EVM implementations.
* it should be possible to create an account with a handler defined in the same compilation unit as our currently executing code just be referring to its Rust or Go type rather than some string
* a handler ID should represent an immutable piece of code, so any migration or upgrade should involve changing the handler ID to a new handler ID, even if it's just a version bump

Possible approaches:
* in Rust, we can use `module_path!()` to get the module path including the crate name, and in Go we can get the package path by reflection
* for SDK modules, we currently use versioned protobuf files to create a module config message, and we should be bumping the version every time we change the module, but I think we're not doing that
* for Wasm, we can use the format `wasm:<hash>:<handler_id>` where `hash` is the hash of the Wasm code and `handler_id` is some handler ID within the Wasm code (maybe coming from a Rust `module_path!()` plus type name in a macro)
* the EVM uses addresses to identify contracts, and all contracts are basically singletons and immutable, so we can just use the format `evm:<address>` and such accounts would never be upgradeable

We can divide the complexity of this problem into two parts: native handlers and VM handlers.
For VM handlers, the problem is simple because we can simply use the hash, and it must be the same for
all nodes even if they use different VM implementations.

For native handlers, if we just took a naive approach of doing something like `rust:<module_path!()>:<type_name>`, then
we have the following problems:
* we can't implement the same handler in Go and Rust, or load the same handler alternatively as a dynamic
library, compiled into the binary or loaded via Wasm
* there is no version information encoded into the handler ID, so if someone just swaps out a binary with
the same handler, then we need to create a separate upgrade path

If we had user-assigned names for Native handlers (like we do with versioned protobuf types), then
we could make the implementation independent of the language and the compilation unit, but then we have these
issues:
* it's harder to refer to a handler in the same compilation unit
* people will probably update binaries without updating the handler ID, so we may need a separate "native" migration path

## Encoding