# ADR 030: Typed Events

## Changelog

- 28-Sept-2020: Initial Draft

## Authors

- Anil Kumar (@anilcse)
- Jack Zampolin (@jackzampolin)
- Adam Bozanich (@boz) 

## Status

Proposed

## Context

Currently in the SDK, events are not properly organized (event data is defined in the handlers for each message) and are prone to inconsistencies. This makes customizing events difficult. In addition, they are difficult to consume. This proposal focuses on updating the events to use **typed events** in each module such that emiting and subscribing to events will be much more egrnomic. This will make use of the SDK to build event driven processes much easier and enable rapid development of things like relayers and automated transaction bots for SDK based chains. These types of bots enable will enable easy building of new features at exchanges, wallets, explorers, and defi protocols. IBC especially will benefit from this proposal.

The end of this proposal contains a detailed example of how to consume events after this refactor.

## Decision

__Step-1__: Declare event types for `sdk.Msg`s a module implements using the typed event interface: `sdk.ModuleEvent`. We first need to define this interface and supporting types.

```go
// types/events.go

// ModuleEvent is the interface that all message events will implement
type ModuleEvent interface {
    Context()   BaseModuleEvent
    ABCIEvent() Event
}

// BaseModuleEvent contains information about the Module and Action of an event
type BaseModuleEvent struct {
    Module string
    Action string
}
```

The `BaseModuleEvent` struct will be used for basic context about the event. It aids in routing events to their 
`Module` and `Action` specific event parsers.

__Step 2__:  Implement additional functionality in the `types` package: utility functions and a parser to route the event to its proper module.

When we subscribe to emitted events on the tendermint websocket, they are emitted in the form of an `abci.Event`. The parser will process this event using `sdk.NewSDKEvent(abci.Event)` to enable passing of the processed event to the proper module.

```go
// types/events.go

// SDKEvent contains the string representation of the event and the module information
type SDKEvent struct {
    Sev  StringEvent
    Base BaseModuleEvent
}

// NewSDKEvent parses abci.Event into an sdk.SDKEvent
func NewSDKEvent(bev abci.Event) (ev SDKEvent, err error) {
    sev := StringifyEvent(bev)
    module, err := GetEventString(sev.Attributes, AttributeKeyModule)
    if err != nil {
        return
    }
    action, err := GetEventString(sev.Attributes, AttributeKeyAction)
    if err != nil {
        return
    }
    return SDKEvent{sev, BaseModuleEvent{module, action}}, nil
}

// GetEventString take sdk attributes, key and returns value for that key. 
func GetEventString(attrs []Attribute, key string) (string, error) {
    for _, attr := range attrs {
        if attr.Key == key {
            return attr.Value, nil
        }
    }
    return "", fmt.Errorf("not found")
}

// GetEventUint64 take sdk attributes, key and returns uint64 value. 
// Returns error incase of failure.
func GetEventUint64(attrs []Attribute, key string) (uint64, error) {
    sval, err := GetEventString(attrs, key)
    if err != nil {
        return 0, err
    }
    return strconv.ParseUint(sval, 10, 64)
}

// Other type functions for use in the individual module parsers
// e.g. func GetEventFloat64(attrs []Attribute, key) (float64, error) {}
```

__Step-3__: Add `AppModuleBasic.ParseEvent` and define `app.BasicManager.ParseEvent`:

A `ParseEvent` function will need to be added to the `sdk.AppModuleBasic` interface.

```go
type AppModuleBasic interface {
    ...
    ParseEvent(ev sdk.SDKEvent) (sdk.ModuleEvent, error)
    ...
}

// ParseEvent takes an sdk.SDKEvent and returns the module specific sdk.ModuleEvent
func (bm BasicManager) ParseEvent(ev sdk.SDKEvent) (sdk.ModuleEvent, error) {
    for m, b := range bm {
        if m == ev.Base.Module {
            return b.ParseEvent(cdc)
        }
    }
    return nil, fmt.Errorf("failed to parse event")
}
```

__Step-4__: Define typed events for msgs in `x/<module>/types/events.go`:

For example, let's take `MsgSubmitProposal` of `gov` module and implement this event's type.

```go
// x/gov/types/events.go
func NewEventSubmitProposal(from sdk.Address, id govtypes.ProposalID, proposal govtypes.TextProposal) EventSubmitProposal {
    return EventSubmitProposal{
        ID:          id,
        FromAddress: from,
        Proposal:    proposal,
    }
}

type EventSubmitProposal struct {
    FromAddress   AccAddress
    ID            ProposalID
    Proposal      types.TextProposal
}

func (ev EventSubmitProposal) Context() sdk.BaseModuleEvent {
    return BaseModuleEvent{
        Module: "gov",
        Action: "submit_proposal",
    }
}

func (ev EventSubmitProposal) ABCIEvent() sdk.Event {
    return types.NewEvent("cosmos-sdk-events",
        sdk.NewAttribute(sdk.AttributeKeyModule, ev.Context().Module),
        sdk.NewAttribute(sdk.AttributeKeyAction, ev.Context().Action),
        sdk.NewAttribute("from", ev.FromAddress.String()),
        sdk.NewAttribute("title", ev.Proposal.Title.String()),
        sdk.NewAttribute("description", ev.Proposal.Description.String()),
    )
}
```

__Step-5__: Define `ParseEvent` for each module in their respective `x/<module>/module.go`:

```go
// x/gov/module.go

// ParseEvent turns an sdk.SDKEvent into the gov specific event type and error if any occurred
func (AppModuleBasic) ParseEvent(ev sdk.SDKEvent) (sdk.ModuleEvent, error) {
    if ev.Sev.Type != sdk.EventTypeMessage {
        return nil, fmt.Errorf("unknown message type")
    }

    if ev.Base.Module != ModuleName {
        return nil, fmt.Errorf("wrong module: %s not %s", ev.Base.Module, ModuleName)
    }
    
    switch ev.Base.Action {
    case "submit_proposal":
        addr, err := sdk.GetEventString(ev.Sev.Attributes, "from")
        if err != nil {
            return nil, err
        }
        proposalId, err := sdk.GetEventUint64(ev.Sev.Attributes, "proposal_id")
        if err != nil {
            return nil, err
        }
        proposal, err := parseProposalFromEvent(ev.Sev.Attributes, "id")
        if err != nil {
            return nil, err
        }
        from, err := sdk.AccAddressFromBech32(addr)
        if err != nil {
            return nil, err
        }
        return NewEventSubmitProposal(from, proposalId, proposal), nil
    case "proposal_deposit":
        // TODO: Implement
    case "submit_proposal":
        // TODO: Implement
    case "proposal_deposit":
        // TODO: Implement
    case "proposal_vote":
        // TODO: Implement
    case "inactive_proposal":
        // TODO: Implement
    case "active_proposal":
        // TODO: Implement
    default:
        return nil, fmt.Errorf("unsupported event type for gov")
    }
}

// parseProposalFromEvent returns the TextProposal from []sdk.Attributes
func parseProposalFromEvent(attrs []sdk.Attribute) ([]byte, error) {
    description, err := sdk.GetEventString(attrs, "description")
    if err != nil {
        return govtypes.TextProposal{}, err
    }

    title, err := sdk.GetEventString(attrs, "title")
    if err != nil {
        return govtypes.TextProposal{}, err
    }

    return govtypes.TextProposal{
        Title:        title,
        Description:  description,
    }, nil
}
```

__Step-6__: Refactor event emission to use the types created:

Emiting events is similar to the current method:

```go
// x/gov/handler.go
func handleMsgSubmitProposal(ctx sdk.Context, keeper keeper.Keeper, msg types.MsgSubmitProposalI) (*sdk.Result, error) {
    ...
    types.Context.EventManager().EmitEvent(
        NewEventSubmitProposal(fromAddress, id, proposal).ABCIEvent(),
    )
    ...
}
```

#### How to subscribe to these typed events in `Client`

> NOTE: Full code example below

Users will be able to subscribe using `client.Context.Client.Subscribe` and consume events which are emitted using `EventHandler`s.

Akash Network has built a simple [`pubsub`](https://github.com/ovrclk/akash/blob/master/pubsub/bus.go). This can be used to subscribe to `abci.Events` and [publish](https://github.com/ovrclk/akash/blob/master/events/publish.go#L21) them as typed events.

Please see the below code sample for more detail on this flow looks for clients.

## Consequences

### Positive

* Improves consistency of implementation for the events currently in the sdk
* Provides a much more ergonomic way to handle events and facilitates writing event driven applications
* This implementation will support a middleware ecosystem of `EventHandler`s

### Negative

* Requires a substantial amount of additional code in each module. For new developers and chains, this can be 
partially mitigaed by code generation in [`starport`](https://github.com/tendermint/starport).

## Detailed code example of publishing events

This ADR also proposes adding affordances to emit and consume these events. This way developers will only need to write
`EventHandler`s which define the actions they desire to take. 

```go
// EventEmitter is a type that describes event emitter functions
// This should be defined in `types/events.go`
type EventEmitter func(context.Context, client.Context, ...EventHandler) error

// EventHandler is a type of function that handles events coming out of the event bus
// This should be defined in `types/events.go`
type EventHandler func(sdk.ModuleEvent) error

// Sample use of the functions below
func main() {
    ctx, cancel := context.WithCancel(context.Background())

    if err := TxEmitter(ctx, client.Context{}.WithNodeURI("tcp://localhost:26657"), SubmitProposalEventHandler); err != nil {
        cancel()
        panic(err)
    }

    return
}

// SubmitProposalEventHandler is an example of an event handler that prints proposal details
// when any EventSubmitProposal is emitted. 
func SubmitProposalEventHandler(ev sdk.ModuleEvent) (err error) {
    switch event := ev.(type) {
    // Handle governance proposal events creation events
    case govtypes.EventSubmitProposal:
        // Users define business logic here e.g.
        fmt.Println(ev.FromAddress, ev.ID, ev.Proposal)
        return nil
    default:
        return nil
    }
}

// TxEmitter is an example of an event emitter that emits just transaction events. This can and 
// should be implemented somewhere in the SDK. The SDK can include an EventEmitters for tm.event='Tx' 
// and/or tm.event='NewBlock' (the new block events may contain typed events) 
func TxEmitter(ctx context.Context, cliCtx client.Context, ehs ...EventHandler) (err error) {
    // Instantiate and start tendermint RPC client
    client, err := cliCtx.GetNode()
    if err != nil {
        return err
    }

    if err = client.Start(); err != nil {
        return err
    }

    // Start the pubsub bus
    bus := pubsub.NewBus()
    defer bus.Close()

    // Initialize a new error group
    eg, ctx := errgroup.WithContext(ctx)

    // Publish chain events to the pubsub bus
    eg.Go(func() error {
        return PublishChainTxEvents(ctx, client, bus, simapp.ModuleBasics)
    })

    // Subscribe to the bus events
    subscriber, err := bus.Subscribe()
    if err != nil {
        return err
    }

	// Handle all the events coming out of the bus
	eg.Go(func() error {
        var err error
        for {
            select {
            case <-ctx.Done():
                return nil
            case <-subscriber.Done():
                return nil
            case ev := <-subscriber.Events():
                for _, eh := range ehs {
                    if err = eh(ev); err != nil {
                        break
                    }
                }
            }
        }
        return nil
	})

	return group.Wait()
}

// PublishChainTxEvents events using tmclient. Waits on context shutdown signals to exit.
func PublishChainTxEvents(ctx context.Context, client tmclient.EventsClient, bus pubsub.Bus, mb module.BasicManager) (err error) {
    // Subscribe to transaction events
    txch, err := client.Subscribe(ctx, "txevents", "tm.event='Tx'", 100)
    if err != nil {
        return err
    }

    // Unsubscribe from transaction events on function exit
    defer func() {
        err = client.UnsubscribeAll(ctx, "txevents")
    }()

    // Use errgroup to manage concurrency
    g, ctx := errgroup.WithContext(ctx)

    // Publish transaction events in a goroutine
    g.Go(func() error {
        var err error
        for {
            select {
            case <-ctx.Done():
                break
            case ed := <-ch:
                switch evt := ed.Data.(type) {
                case tmtypes.EventDataTx:
                    if !evt.Result.IsOK() {
                        continue
                    }
                    // range over events, parse them using the basic manager and 
                    // send them to the pubsub bus
                    for _, abciEv := range events {
                        sdkEv, err := sdk.NewSDKEvent(abciEv)
                        if err != nil {
                            return err
                        }
                        moduleEvent, err := mb.ParseEvent(abciEv)
                        if err != nil {
                            return er
                        }
                        if err := bus.Publish(moduleEvent); err != nil {
                            bus.Close()
                            return
                        }
                        continue
                    }
                }
            }
        }
        return err
	})

    // Exit on error or context cancelation
    return g.Wait()
}
```

## References
- [Event types for a module](https://github.com/ovrclk/akash/blob/master/x/deployment/types/event.go#L24)
- [Emit Events](https://github.com/ovrclk/akash/blob/master/x/deployment/keeper/keeper.go#L129)
- [Publish Custom Events via a bus](https://github.com/ovrclk/akash/blob/master/events/publish.go#L19-L58)
- [Consuming the events in `Client`](https://github.com/jackzampolin/deploy/blob/master/cmd/event-handlers.go#L57)
