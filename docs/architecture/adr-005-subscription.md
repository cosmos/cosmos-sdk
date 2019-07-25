# ADR 5: Subscription

## Changelog


## Context

Cosmos-SDK should allow developers to choose from a variety of payment models for their application. One of the most popular payment models for Internet applications, so it is natural to provide this payment model as a module for sdk users.

For fuller context around this issue: see [\#4642](https://github.com/cosmos/cosmos-sdk/issues/4642)

## Decision

Create Subscription module in SDK so that users can subscribe to various subscription services. Subscription collectors will send messages to collect subscription payments from users who have due subscriptions. Subscriptions are paid out directly from a user account. If the user does not have enough funds to pay for the subscription is inactivated. Users can define maximum limits on how many periods the subscription is valid for; if this limit elapses before a subscription renewal, the subscription is invalidated.

## Status

#### Proposed

## Implementation Changes

Introduce the following messages to `x/subscription` module:

```go
// CreateSubscriptionMsg allows Service Provider to create new subscription service.
type CreateSubscriptionMsg struct {
    Name      string         // unique, human-readable name for subscription service
    Amount    sdk.Coins      // amount to be collected for each subscription period
    Period    time.time      // duration of subscription period
    Collector sdk.AccAddress // address that will collect subscription payments
}
```

```go
// SubscribeMsg allows subscriber account to subscribe to provided service
// If subscription to the service from user already exists, this msg is treated as renewal.
// If msg.Limit != -1, then subscription.Limit += msg.Limit. Else, subscription.Limit = -1 (unlimited)
type SubscribeMsg struct {
    Name       string         // name of service to subscribe to
    Subscriber sdk.AccAddress // address of subscriber
    Limit      int64          // Maximum number of periods that subscription remains active. Limit = -1 implies no limit
}
```

```go
// UnsubscribeMsg allows subscriber to inactivate an active subscription
type UnsubscribeMsg struct {
    Name       string         // name of service to unsubscribe to
    Subscriber sdk.AccAddress // address of subscriber
}
```

```go
// CollectMsg allows a Collector for a service to collect payments on due subscriptions that are processed off a FIFO queue.
type CollectMsg struct {
    Name      string         // name of subscription service to collect payments from
    Collector sdk.AccAddress // address that will collect payments
    Limit     int64          // maximum number of items to process in FIFO duequeue. If Limit = -1, try to process all due subscriptions
}
```

Create a new store with the following key-values:

`Address => []SubscriptionID // List of active/inactive subscriptions owned by the users`

`Metadata:{Name} => SubscriptionMetaData`

`DueQueue:{Name} => []SubscriptionID // FIFO queue of subscriptions. CONTRACT: All due subscriptions exist before all undue subscriptions. All subscriptions in DueQueue are active`

`SubscriptionID => Subscription`

```go
// SubscriptionID is a unique identifier to a subscription struct
// hash(Name+Address)
type SubscriptionID []byte
```

```go
// Subscription Metadata contains information necessary for processing subscriptions
type Metadata struct {
    Amount    sdk.Coins      // amount to be collected for each subscription period
    Period    time.time      // duration of subscription period
    Collector sdk.AccAddress // address that will collect subscription payments
}
```

```go
// Represents a user's subscription to a service
type Subscription {
    Name       string
    Subscriber sdk.AccAddress
    Limit      int64
    LastPaid   time.Time
    Active     bool
}
```

Message Handling:

Note much of this logic may be done inside a keeper. This is a outline of how msgs will be handled.

```go
func HandleCreateSubscriptionMsg(ctx sdk.Context, msg CreateSubscriptionMsg) {
    if SubscriptionExists(msg.Name) {
        return DuplicateSubscriptionErr
    }
    StoreMetadata(ctx, msg.Name, msg.Amount, msg.Period, msg.Collector) // Store metadata in store under key "Metadata:msg.Name"
    InitializeDueQueue(ctx, msg.Name) // Initialize empty queue and store under key "DueQueue:Name"
}
```

```go
func HandleSubscribeMsg(ctx sdk.Context, msg SubscribeMsg) {
    if !SubscriptionExists(msg.Name) {
        return NoSuchSubscriptionErr
    }
    subscriptionID := hash(msg.Name|msg.Subscriber)
    // check if user already has subscribe to this service
    // if so renew the subscription
    if subscription := GetUserSubscriptions(msg.Subscriber, msg.Name); subscription != nil {
        if msg.Limit == -1 {
            subscription.Limit = -1
        } else {
            subscription.Limit += msg.Limit
        }
        // set active to true
        subscription.Active = true
        StoreSubscription(subscriptionID, subscription)
        return nil
    }
    metadata := GetMetaData(ctx, msg.Name)
    // subscribeMsg will pay for the first period
    err := bank.Send(msg.Subscriber, metadata.Collector, metadata.Amount)
    if err != nil {
        return err
    }
    subscription := Subscription{
        Name:       metadata.Name,
        Subscriber: msg.Subscriber,
        Limit:      msg.Limit,
        LastPaid:   ctx.BlockTime,
        Active:     true
    }
    StoreSubscription(subscriptionID, subscription)
    // adds subscription to list of user subscriptions
    AppendToUserSubscriptions(subscriptionID)
    // push to back of DueQueue for this service
    PushToBackOfDueQueue(subscriptionID)
}
```

```go
func HandleUnsubscribeMsg(ctx sdk.Context, msg UnsubscribeMsg) {
    subscription := GetUserSubscriptions(msg.Subscriber, msg.Name)
    if subscription != nil {
        return nil
    }
    subscription.Active = false
    subscription.Limit = 0 // to ensure that renewal updates limit correctly
    subscriptionID := hash(msg.Name|msg.Subscriber)
    StoreSubscription(subscriptionID, subscription)
}
```

```go
func HandleCollectMsg(ctx sdk.Context, msg CollectMsg) {
    if !SubscriptionExists(msg.Name) {
        return NoSuchSubscriptionErr
    }
    metadata := GetMetaData(ctx, msg.Name)
    if msg.Collector != metadata.Collector {
        return InvalidCollectorErr
    }
    queue := GetDueQueue(msg.Name)
    for i := 0; i < msg.Limit; i++ {
        // top of the queue is "due"
        subscription := GetSubscription(queue[0])
        if subscription.LastPaid + metadata.Period >= ctx.BlockTime {
            err := bank.Send(subscription.Subscriber, metadata.Collector, metadata.Amount)
            if err != nil {
                // insufficient funds. inactivate subscription
                subscription.Active = false
                subscription.Limit = 0
            } else {
                // update subscription after successful payment
                subscription.LastPaid = ctx.BlockTime
                if subscription.Limit != -1 {
                    subscription.Limit--
                }
            }
            // store updated subscription
            StoreSubscription(subscriptionID, subscription)
            // push paid subscription back to end of queue
            PushToBack(queue, subscriptionID)
        } else {
            // processed all due subscriptions
            return nil
        }
    }
}
```


## Consequences

### Positive

- Adds a useful module to core-sdk that can be used by many future applications

### Negative

- One more module in `x/` folder

## References

- [Issue \#4642](https://github.com/cosmos/cosmos-sdk/issues/4642)

