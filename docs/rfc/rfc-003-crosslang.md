# RFC 003: Account, Module, Message Model

## Changelog

* {date}: {changelog}

## Background

> The next section is the "Background" section. This section should be at least two paragraphs and can take up to a whole
> page in some cases. The guiding goal of the background section is: as a newcomer to this project (new employee, team
> transfer), can I read the background section and follow any links to get the full context of why this change is  
> necessary?
>
> If you can't show a random engineer the background section and have them acquire nearly full context on the necessity
> for the RFC, then the background section is not full enough. To help achieve this, link to prior RFCs, discussions, and
> more here as necessary to provide context so you don't have to simply repeat yourself.


## Proposal

> The next required section is "Proposal" or "Goal". Given the background above, this section proposes a solution.
> This should be an overview of the "how" for the solution, but for details further sections will be used.

The base layer framework component for the cross-language framework is the **hypervisor** which specifies the meaning of the core concepts of **account**, **message**, **code environment**, **module** and **gas**.

An **account** is defined as having:
* a unique **account address**
* a code handler which allows it to execute **messages**

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
