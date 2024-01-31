# RFC 006: Handlers

## Changelog

* January 26, 2024: Initialized

## Background

The Cosmos SDK has a very powerful and flexible module system, it has been tested and proven to be very good in production.
The design of how messages are handled is built around proto buf services and Grpc. This design was proposed and 
implemented during the time we migrated from amino to protocol buffers. This design has fullfilled the needs of users
today. While this design is useful it has caused a elevated learning curve to be adopted by users. Today, these services 
are the only way to write a module. This RFC proposes a new design that simplifies the design and enables new use cases
we are seeing today. 

Taking a step back we have seen the emergence of rollups and proving technologies. These technologies enable new use cases
and new methods of achieveing various goals. When we look at things like proving we look to [Tinygo](https://tinygo.org/). When we have
attempted to use tinygo with existing modules we have run into a hiccup, the use of [Grpc](https://github.com/tinygo-org/tinygo/issues/2814) within modules. This has
led us to look at a design which would allow the usage of tinygo and other technologies.

We looked at Tinygo for our first target in order to compile down down to a 32 bit environment which could be used with
things like [Risc-0](https://www.risczero.com/), [Fluent](https://fluentlabs.xyz/) and other technologies. When speaking with the teams behind these technologies
we found that they were interested in using the Cosmos SDK but were unable to due to being unable to use Tinygo or the
Cosmos SDK go code in a 32 bit environment.

The Cosmos SDK team has been hard at work over the last few months designing and implementing a modular core layer, with
the idea that proving can be enabled later on. This design allows us to push the design of what can be done with the
Cosmos SDK to the next level. In the future when we have proving tools and technologies integrated parts of the new core
layer will be able to be used in conjunction with proving technologies without the need to rewrite the stack. 


## Proposal

This proposal is around enabling modules to be compiled to an environment in which they can be used with Tinygo and/or
different proving technologies.

> Note the usage of handlers in modules is optional, modules can still use the existing design. This design is meant to
> be a new way to write modules, with proving in mind, and is not meant to replace the existing design.

### Pre and Post Message Handlers 

In the Cosmos SDK, there exists hooks on messages and execution of function calls. Separating the two we will focus on
message hooks. When a message is implemented it can be unknown if others will use the module and if a message will need
hooks. When hooks are needed before or after a message, users are required to fork the module. This is not ideal as it 
leads to a lot of forks of modules and a lot of code duplication.

Pre and Post message handlers solve this issue. Where we allow modules to register listeners for messages in order to
execute something before and/or after the message. Although hooks can be bypassed by doing keeper function calls, we can
assume that as we shift the communication design of the SDK to use messages instead of keeper calls it will be safe to
assume that the bypass surface of hooks will reduce to zero.

If an application developer would like to check the sender of funds before the message is executed they can
register a pre message handler. If the message is called by a user the pre message handler will be called with the custom
logic. If the sender is not allowed to send funds the premessage handler can return an error and the message will not be
executed.

A module can register handlers for any or all message(s), this allows for modules to be extended without the need to fork.

A module will implement the below for a premessage hook:

```golang
package core_appmodule

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

type PreMsgHandlerRouter interface {
	// RegisterGlobalPreMsgHandler will register a pre msg handler that hooks before any message executes.
	// Handler will be called before ANY message executes.
	RegisterGlobalPreMsgHandler(handler func(ctx context.Context, msg protoiface.MessageV1) error)
	// RegisterPreMsgHandler will register a pre msg handler that hooks before the provided message
	// with the given message name executes. Handler will be called before the message is executed
	// by the module.
	RegisterPreMsgHandler(msgName string, handler func(ctx context.Context, msg protoiface.MessageV1) error)
}

type HasPreMsgHandler interface {
	RegisterPreMsgHandler(router PreMsgHandlerRouter)
}
```

A module will implement the below for a postmessage hook:

```go
package core_appmodule

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

type PostMsgHandlerRouter interface {
	// RegisterGlobalPostMsgHandler will register a post msg handler that hooks after any message executes.
	// Handler will be called after ANY message executes, alongside the response.
	RegisterGlobalPostMsgHandler(handler func(ctx context.Context, msg, msgResp protoiface.MessageV1) error)
	// RegisterPostMsgHandler will register a pre msg handler that hooks after the provided message
	// with the given message name executes. Handler will be called after the message is executed
	// by the module, alongside the response returned by the module.
	RegisterPostMsgHandler(msgName string, handler func(ctx context.Context, msg, msgResp protoiface.MessageV1) error)
}

type HasPostMsgHandler interface {
	RegisterPostMsgHandler(router PostMsgHandlerRouter)
}
```

We note the following behaviours:
- A pre msg handler returning an error will yield to a state transition revert.
- A post msg handler returning an error will yield to a state transition revert.
- A post msg handler will not be called if the execution handler (message handler) fails.
- A post msg handler will be called only if the execution handler succeeds alongside the response provided by the execution handler.

### Message and Query Handlers

Similar to the above design, message handlers will allow the application developer to replace existing Grpc based services 
with handlers. This enables the module to be compiled down to tinygo, and abandon the gRPC dependency. As mentioned
upgrading the modules immediately is not mandatory, module developers can do so in a gradual way.

For message handlers we propose the introduction of the following core/appmodule interfaces and functions:

```go
package core_appmodule

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

type MsgHandlerRouter interface {
	RegisterMsgHandler(msgName string, handler func(ctx context.Context, msg protoiface.MessageV1) (msgResp protoiface.MessageV1, err error))
}

type HasMsgHandler interface {
	RegisterMsgHandlers(router MsgHandlerRouter)
}

// RegisterMsgHandler is a helper function to retain type safety when creating handlers, so we do not need to cast messages.
func RegisterMsgHandler[Req, Resp protoiface.MessageV1](router MsgHandlerRouter, handler func(ctx context.Context, req Req) (resp Resp, err error)) {
	// impl detail
}
```

Example

```go
package bank 

func (b BankKeeper) Send(ctx context.Context, msg bank.MsgSend) (bank.MsgSendResponse, error) {
	// logic
}

func (b BankModule) RegisterMsgHandlers(router core_appmodule.MsgHandlerRouter) {
	// the RegisterMsgHandler function takes care of doing type casting and conversions, ensuring we retain type safety
	core_appmodule.RegisterMsgHandler(router, b.Send) 
}

```

This change is fully client API compatible, and state machine compatible as, even if we were using gRPC, messages were
routed using the message type name and not the gRCP method name.

We apply the same principles of MsgHandlers to QueryHandlers, by introducing a new core/appmodule interface:

```go
package core_appmodule

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

type QueryHandlerRouter interface {
	RegisterQueryHandler(msgName string, handler func(ctx context.Context, msg protoiface.MessageV1) (msgResp protoiface.MessageV1, err error), aliases ...string)
}

type HasQueryHandler interface {
	RegisterQueryHandlers(router QueryHandlerRouter)
}

// RegisterQueryHandler is a helper function to retain type safety when creating handlers, so we do not need to cast messages.
func RegisterQueryHandler[Req, Resp protoiface.MessageV1](router QueryHandlerRouter, handler func(ctx context.Context, req Req) (resp Resp, err error)) {
	// impl detail
}

```

The difference between gRPC handlers and query handlers is that we expect query handlers to be deterministic and usable
in consensus by other modules. Non consensus queries should be registered outside of the state machine itself, and we will
provide guidelines on how to do so with serverv2.


### Consensus Message Handlers


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
