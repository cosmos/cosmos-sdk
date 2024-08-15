# RFC 003: Account/Module Execution Model

## Changelog

* 2024-08-09: Reworked initial draft (previous work was in https://github.com/cosmos/cosmos-sdk/pull/15410)

## Background

## Proposal

We propose a conceptual and formal model for defining **accounts** and **modules** which can interoperate with each other through **messages** in a cross-language environment.

### Core Concepts

We start with the conceptual definition of some core concepts from the perspective of a developer trying to write code for a module or account. The formal representation of these concepts in a particular coding environment may vary depending on that environment, but the essence should remain more or less the same. Formal system-wide definitions will be given later in the section on the **hypervisor** and **virtual machines**.

An **account** is defined as having:
* a unique **address**
* an **account handler** which is some code which can process **messages** and send **messages** to other **accounts**
* some optional **account config** data

An **address** is defined as a variable-length byte array of up to 255 bytes, although this may be subject to change.

A **message** is defined as a tuple of:
* a **message name** 
* and **message data**

When a **message** is sent to an **account**'s code handler, that handler will receive:
* the **address** of the **account** (its own address)
* the **address** of the account sending the message (the caller), which will be empty if the message is a query
* the **message name**
* the **message data**
* a **gas limit**
* **account config** data

The handler can then execute some code and return a response or an error.

To send a **message** to another **account**, the caller must specify:
* the **message name**
* the **message data**
* a **gas limit**
* and optionally, the **address** of the **account** to send the message to

There is a special class of **message name**'s known as **module messages**,
where the caller should omit the **address** of the receiving **account**.
The routing framework can look up the **address** of the receiving **account** based on the **message name** of a **module message**.

**Accounts** which define handlers for **module messages** are known as **modules**.

**Account**s generally also have some mutable state, but within this specification, state is expected to be handled by some special state **module** which is defined by separate specifications.

**Account handlers** should provide metadata about themselves and all the **messages that they handle.
In the current design, the only part of the **message** metadata
which is standardized at this level is an enum value, **volatility**,
which describes the behavior of the handler with respect to state.
This `volatile` enum can have one of the following values:
* `volatile`: the handler can have side effects and send `volatile`, `radonly` or `pure` messages to other accounts. Such handlers are expected to both read and write state.
* `readonly`: the handler cannot cause effects side effects and can only send `readonly` or `pure` messages to other accounts. Such handlers are expected to only read state.
* `pure`: the handler cannot cause any side effects and can only call other pure handlers. Such handlers are expected to neither read nor write state.

For the time being, further standardization around message metadata is expected to be done at other levels of the stack,
and as far as this specification is concerned, message metadata should be considered opaque bytes.

### Account Lifecycle

**Accounts* can be created, destroyed and migrated to new **account handlers**.

**Account handlers** can define code for the following special **message name**'s:
* `on_create`: called when an account is created with **message data** containing arbitrary initialization data
* `on_destroy`: called when an account is destroyed with **message data** containing arbitrary destruction data
* `on_migrate`: called when an account is migrated to a new code handler. Such handlers receive structured **message data** specifying the old code handler so that the account can perform migration operations or return an error if migration is not possible.

### Hypervisor and Virtual Machines

A special module known as the **hypervisor** manages:
* the mapping of **account address** to **account handler**
* the mapping of **message name** to **account address** for **module messages**
* the creation, destruction and migration of accounts
* storage of state **account config** data
* **virtual machines** which run **account handlers**
* routing of **messages** to **account handlers** for both internal and external callers
* loading of module configuration data (app config)

Each **account handler** runs inside in a specific code environment which for simplicity we will refer to as a **virtual machine**.
These code environments may or may not be sand-boxed **virtual machine**s in the strictest sense.
For example, one such code environment may be native golang code while another may be an actual WASM virtual machine.
However, for consistency we will refer to these all as **virtual machines** 
because the **hypervisor** will interact with them in the same way.

Each **virtual machine** is expected to expose a `handle` function which takes the following parameters:
* **handler id**, which is a byte array that the **virtual machine** maps to a specific **account handler**
* **account address**
* caller **account address**
* **message name**
* **message data**
* **gas limit**

`handle` returns an optional message response, the amount of gas consumed, or an error if the handler failed.

Each **virtual machine** environment receives a callback function
`invoke` which allows it to send messages to other accounts.
`invoke` takes the following parameters:
* **message name**
* target **account address**, which may be empty if the **message name** refers to a **module message**
* **message data**
* **gas limit**

`invoke` returns the same data as `handle`.

**Virtual machine**s should also implement a `describe` function which returns metadata about the code handlers it supports. `describe` takes the **handle id** as an input parameter and returns an array of **message** handler metadata containing:
- the **message name**
- its **volatility** enum value 
- opaque bytes of additional metadata

The **hypervisor** module itself handles the following **module messages**:
* `create(code id, account address?)`: creates a new account in the specified code environment with the specified code id and optional pre-defined account address (if not provided, a new address is generated). The `on_create` message is called if it is implemented by the account.
* `destroy(account address)`: deletes the account with the specified address and calls the `on_destroy` message if it is implemented by the account.
* `migrate(account address, new code id, miration data)`: migrates the account with the specified address to the new code environment and code id. The `on_migrate` message must be implemented by the new code and must not return an error for migration to succeed. `migrate` can only be called by the account itself (or by the app outside normal execution flow).
* `force_migrate(account address, new code id, init data, destroy data)`: this can be used when no `on_migrate` handler can perform a proper migration to the new code. In this case, `on_destroy` will be called on the old event, its state will be cleared, and `on_create` will be called on the new code. This is a destructive operation and should be used with caution.

### Module Lifecycle & Messages

For legacy purposes, **modules** have specific lifecycles and **module messages** have special semantics.

When the **hypervisor** loads each **virtual machine** environment, it will ask the **virtual machine** to describe all the 

--------------------------------------------

Every **account** runs in a **code environment**. The **hypervisor** contains a stateful mapping from `account address -> (code environment, code id)`. The **hypervisor** executes messages by specifying the **account address** and **message name**.

**Accounts** can register themselves as the default handler for a given message name. When this occurs a **message** can be invoked without knowing the address of the handling module - instead the **hypervisor** knows the mapping from **message name** to default account address. **Accounts** which register default message handlers are known as **modules**. Thus, the **hypervisor** also contains a stateful mapping from `message name -> account address`. TODO: can other accounts handle messages with the same name as messages that have a default handler, i.e. are "modules messages"?? 

Every **code environment** must expose a `handle` function: which takes a `(code id, message data, account address, caller account address?, gas limit)` as input parameters and returns an optional response plus gas consumed. The `handle` function is the entry point for executing messages. `code id`s have a format defined by their code environment, although some common standards may emerge. `message data` is expected to include the message name embedded for routing purposes, and the code environment is expected to parse this data and route it to the appropriate handler. The `caller account address` is specified except when the message is a query (TODO: how to specify queries? - either by message name or with a separate query handler). 

Every code environment receives the following callback functions:
* `invoke(account address, message data, gas limit)`: sends a message to another account specified by its address
* `invoke_default(message name, message data, gas limit)`: sends a message to the account registered with the default handler for the given message name, if one exists
* `register_default_handler(message name)`: registers the account as the default handler for the given message name. This will fail if two accounts try to register as the default handler for the same message name

**Gas limit** parameters are an integer which specifies the maximum number of gas units that may be consumed during the execution of the handler before the code environment returns an error. When each handler returns it should return the amount of gas consumed. Each code environment is expected to track execution cost consistently to avoid unbounded execution. The remaining gas should be passed to each nested call to enforce gas limits across the system.

The following special messages are defined at the framework level and can optionally be implemented by accounts:
* `on_create(init data)`: called when an account is created
* `on_destroy(destroy data)`: called when an account is destroyed
* `on_migrate(old code environment, old code id, migration data)`: called when an account is migrated to a new code environment and id. The previous code environment and id should be used to perform migration operations on the old state. If the old state can't be migrated, then the account should return an error.

The **hypervisor** is itself the **root account** and understands the following special messages: 
* `create(code environment, code id, account address?)`: creates a new account in the specified code environment with the specified code id and optional pre-defined account address (if not provided, a new address is generated). The `on_create` message is called if it is implemented by the account.
* `destroy(account address)`: deletes the account with the specified address and calls the `on_destroy` message if it is implemented by the account.
* `migrate(account address, new code environment, new code id, miration data)`: migrates the account with the specified address to the new code environment and code id. The `on_migrate` message must be implemented by the new code and must not return an error for migration to succeed. `migrate` can only be called by the account itself (or by the app outside normal execution flow).
* `force_migrate(account address, new code environment, new code id, init data, destroy data)`: this can be used when no `on_migrate` handler can perform a proper migration to the new code. In this case, `on_destroy` will be called on the old event, its state will be cleared, and `on_create` will be called on the new code. This is a destructive operation and should be used with caution.

Any other specifications regarding the encoding of messages, storage, events, transaction execution or interaction with consensus environments should get specified at a level above the cross-language framework. The cross-language framework is intended to be a minimal specification that allows for the execution of messages across different code environments.

TODO: packet sizes and any details of message data and responses? what size packets are allowed?

TODO: should message names be specified globally? do we have any concept of services (bundles of message handlers)? should any encoding details be specified at this level? 

TODO: how do we deal with pre- and post-handlers that have been specified now in core? are these a hypervisor concern, can they be dealt with at another level, or should they be unsupported?


```go
package hypervisor

type VirtualMachine interface {
	Invoke(HandleArgs) error
	DescribeAccountHandler(handlerId string) (HandlerDescriptor, error)
}

type HandleArgs struct {
	HandlerID string
	Packet MessagePacket
}

type MessagePacket interface{
    AccountAddress() Address
    CallerAddress() Address
    MessageName() string
    MessageData() []byte
    GasLimit() uint64
    AccountConfig() []byte
    ContextToken() []byte
    Param1() []byte
    Param2() []byte
}

type Address = [256]byte

type HandlerDescriptor struct {
	HandlerID string
    MessageDescriptors []MessageDescriptor
	Metadata []byte
}

type MessageDescriptor struct {
	MessageName string
	Volatility  Volatility    
    Metadata    []byte
}

type Volatility uint8

const (
	Volatile Volatility = iota
	ReadOnly
	Pure
)

```

### Message Packet

We specify the format for message packets in Rust to precisely specify memory layout:

```rust
#[repr(packed)]
struct MessagePacket {
    account_address: Address,
    caller_address: Address,
    message_name_len: u8,
    message_name: [u8; 255],
    gas_limit: u64,
    account_config_len: u32,
    account_config: *const u8,
    context_token: [u8; 32],
    // TODO:
    param1_len: u32,
    param1_capacity: u32,
    param1: *mut u8,    
    param2_len: u32,    
    param2_capacity: u32,
    param2: *mut u8,
    message_data_len: u16,
    message_data: [u8; 4096],
}

struct Address {
    len: u8,
    bytes: [u8; 255],
}
```

## Abandoned Ideas (Optional)

> As RFCs evolve, it is common that there are ideas that are abandoned. Rather than simply deleting them from the
> document, you should try to organize them into sections that make it clear they're abandoned while explaining why they
> were abandoned.
>
> When sharing your RFC with others or having someone look back on your RFC in the future, it is common to walk the same
> path and fall into the same pitfalls that we've since matured from. Abandoned ideas are a way to recognize that path
> and explain the pitfalls and why they were abandoned.

## Decision

> This section describes alternative designs to the chosen design. This section
> is important and if an adr does not have any alternatives then it should be
> considered that the ADR was not thought through.

## Consequences (optional)

> This section describes the resulting context, after applying the decision. All
> consequences should be listed here, not just the "positive" ones. A particular
> decision may have positive, negative, and neutral consequences, but all of them
> affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section
> describing these incompatibilities and their severity. The ADR must explain
> how the author proposes to deal with these incompatibilities. ADR submissions
> without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}



### References

> Links to external materials needed to follow the discussion may be added here.
>
> In addition, if the discussion in a request for comments leads to any design
> decisions, it may be helpful to add links to the ADR documents here after the
> discussion has settled.

## Discussion

> This section contains the core of the discussion.
>
> There is no fixed format for this section, but ideally changes to this
> section should be updated before merging to reflect any discussion that took
> place on the PR that made those changes.
