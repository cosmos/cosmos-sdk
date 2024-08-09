# RFC 003: Cross-Language Account Manager Specification

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

The base layer framework component for the cross-language framework is the **account manager** which specifies the meaning of the core concepts of **account**, **message**, **code environment** and **module**.

An **account** is defined as having:
* a unique **account address**
* a code handler which allows it to execute **messages**

Every **account** runs in a **code environment**. The **account manager** contains a stateful mapping from `account address -> (code environment, code id)`. **Messages** in the **account manager** can be executed by specifying either:
1. an **account address** and **message data**, or
2. a **message name** and **message data** (when a default message handler has been registered)

**Accounts** can register themselves as the default handler for a given message name. When this occurs a **message** can be invoked without knowing the address of the handling module - instead the **account manager** knows the mapping from **message name** to default account address. **Accounts** which register default message handlers are known as **modules**. Thus, the account manager also contains a stateful mapping from `message name -> account address`. TODO: can other accounts handle messages with the same name as messages that have a default handler, i.e. are "modules messages"?? 

Every **code environment** must expose a `handle` function: which takes a `(code id, message data, account address, caller account address?)` as input parameters and returns an optional response. The `handle` function is the entry point for executing messages. `code id`s have a format defined by their code environment, although some common standards may emerge. `message data` is expected to include the message name embedded for routing purposes and the code environment is expected to parse this data and route it to the appropriate handler. The `caller account address` is specified except when the message is a query (TODO: how to specify queries? - either by message name or with a separate query handler).

Every code environment receives the following callback functions:
* `invoke(account address, message data)`: sends a message to another account specified by its address
* `invoke_default(message name, message data)`: sends a message to the account registered with the default handler for the given message name, if one exists

The following special messages are defined at the framework level and can optionally be implemented by accounts:
* `on_create()`: called when an account is created
* `on_destroy()`: called when an account is destroyed
* `on_migrate(old code environment, old code id)`: called when an account is migrated to a new code environment

The account manager is itself the **root account** and understands the following special messages: 
* `create(code environment, code id, account address?)`: creates a new account in the specified code environment with the specified code id and optional pre-defined account address (if not provided, a new address is generated). The `on_create` message is called if it is implemented by the account.
* `destroy(account address)`: deletes the account with the specified address and calls the `on_destroy` message if it is implemented by the account.
* `migrate(account address, new code environment, new code id)`: migrates the account with the specified address to the new code environment and code id. The `on_migrate` message is called if it is implemented by the account. `migrate` can only be called by the account itself (or by the app outside of normal execution flow).

Any other specifications regarding the encoding of messages, storage, events, transaction execution or interaction with consensus environments should get specified at a level above the cross-language framework. The cross-language framework is intended to be a minimal specification that allows for the execution of messages across different code environments.

TODO:
* the packet sizes and formats of message data and message response
* how is gas handled? this is likely a core part of the execution framework

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
