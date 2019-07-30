# ADR 5: Subscription

## Changelog


## Context

Cosmos-SDK should allow developers to choose from a variety of payment models for their application. Subscriptions are one of the most popular payment models for Internet applications, so it is natural to provide this payment model as a module for sdk users.

For fuller context around this issue: see [\#4642](https://github.com/cosmos/cosmos-sdk/issues/4642)

## Decision

Create Subscription module in SDK so that users can subscribe to various subscription services. Subscription collectors will send messages to collect subscription payments from users who have due subscriptions. Subscriptions are paid out directly from a user account. 

Service providers can define valid periods their users can choose from (e.g.: 3 months, 6 months, 1 year) and their associated rates when they create the subscription service.

If the user does not have enough funds to pay for the current period, the subscription is inactivated. The user can always refund the account and resubscribe.

Users can define maximum limits on how many periods the subscription is valid for. If this limit elapses before a subscription renewal, the subscription is invalidated. This is to mitigate the common problem of unused subscriptions siphoning off funds from account without user noticing. For example, a user can subscribe to a service with a period of 3 months with a limit of 4. This subscription will automatically expire after a year unless the user manually increases/removes the limit with a second `SubscribeMsg`.

Users can also manually unsubscribe by submitting an `UnsubscribeMsg`

## Status

#### Proposed

## Implementation Changes

Introduce the following messages to `x/subscription` module:

```go
// CreateSubscriptionMsg allows Service Provider to create new subscription service.
// Generate unique serviceID from hash(name|collector)
type CreateSubscriptionMsg struct {
    Name        string          // human-readable name for subscription service
    Description string          // description for what the service provides
    Amounts     []sdk.Coins     // amounts to be collected for each subscription period
    Periods     []time.Duration // allowed duration of subscription periods
    Collector   sdk.AccAddress  // address that will collect subscription payments
}
```

```go
// SubscribeMsg allows subscriber account to subscribe to provided service
// If subscription to the service from user already exists, this msg is treated as renewal.
// If msg.Limit != -1, then subscription.Limit += msg.Limit. Else, subscription.Limit = -1 (unlimited)
type SubscribeMsg struct {
    ServiceID  []byte         // id of service to subscribe to
    Subscriber sdk.AccAddress // address of subscriber
    Period     time.Duration  // Period that subscriber chooses. Must be one of predefined periods in corresponding CreateSubscriptionMsg
    Limit      int64          // Maximum number of periods that subscription remains active. Limit = -1 implies no limit
}
```

```go
// UnsubscribeMsg allows subscriber to inactivate an active subscription
type UnsubscribeMsg struct {
    ServiceID  []byte         // id of service to unsubscribe to
    Subscriber sdk.AccAddress // address of subscriber
}
```

```go
// CollectMsg allows a Collector for a service to collect payments on due subscriptions that are processed off a FIFO queue.
type CollectMsg struct {
    ServiceID []byte         // id of subscription service to collect payments from
    Collector sdk.AccAddress // address that will collect payments
    Limit     int64          // maximum number of items to process in FIFO duequeue. If Limit = -1, try to process all due subscriptions
}
```

Create a new store with the following key-values:

`Address => []SubscriptionID // List of active/inactive subscriptions owned by the users`

`Terms:{ServiceID} => SubscriptionTerms`

`DueQueue:{Collector}{ServiceID}{Period} => LinkedList<SubscriptionID> // FIFO queue of subscriptions for a given service and period. Note if a service allows for multiple periods, each period will maintain a separate queue. CONTRACT: All due subscriptions exist before all undue subscriptions. All subscriptions in DueQueue are active`

`SubscriptionID => Subscription`

```go
// SubscriptionID is a unique identifier to a subscription struct
// hash(Name+Address)
type SubscriptionID []byte
```

```go
// Subscription Terms contains information necessary for processing subscriptions
type Terms struct {
    Name        string         // short, human-readable name of the service
    Description string         // long-form description of the service
    Amount      sdk.Coins      // amount to be collected for each subscription period
    Period      time.Duration  // duration of subscription period
    Collector   sdk.AccAddress // address that will collect subscription payments
}
```

```go
// Represents a user's subscription to a service
type Subscription {
    Name       string
    Subscriber sdk.AccAddress
    Limit      int64
    LastPaid   time.Time
}
```

Message Handling:

Note much of this logic may be done inside a keeper. This is a outline of how msgs will be handled.

```go
func HandleCreateSubscriptionMsg(ctx sdk.Context, msg CreateSubscriptionMsg) {
    serviceID := hash(msg.Name, msg.Collector)
    if SubscriptionExists(serviceID) {
        return DuplicateSubscriptionErr
    }
    StoreTerms(ctx, serviceID, msg.Name, msg.Amount, msg.Period, msg.Collector) // Store Terms in store under key "Terms:serviceID"
    InitializeDueQueues(ctx, msg.Collector, serviceID, msg.Periods) // Initialize a seperate empty queue for each valid period and store under key "DueQueue:Collector:ServiceID:Period"
}
```

```go
func HandleSubscribeMsg(ctx sdk.Context, msg SubscribeMsg) {
    if !SubscriptionExists(msg.ServiceID) {
        return NoSuchSubscriptionErr
    }

    subscriptionID := hash(msg.ServiceID|msg.Subscriber)
    Terms := GetTerms(ctx, msg.ServiceID)

    if msg.Period not in Terms.Periods {
        return InvalidPeriodErr
    }
    
    // check if user already has subscribe to this service
    // if so renew the subscription
    if subscription := GetUserSubscriptions(msg.Subscriber, msg.ServiceID); subscription != nil {
        if msg.Limit == -1 {
            subscription.Limit = -1
        } else {
            subscription.Limit += msg.Limit
        }
        StoreSubscription(subscriptionID, subscription)
        return nil
    }

    // new subscribeMsg will pay for the first period
    err := bank.Send(msg.Subscriber, Terms.Collector, Terms.Amount)
    if err != nil {
        return err
    }

    subscription := Subscription{
        Name:       Terms.Name,
        Subscriber: msg.Subscriber,
        Limit:      msg.Limit,
        LastPaid:   ctx.BlockTime,
    }
    StoreSubscription(subscriptionID, subscription)
    // adds subscription to list of user subscriptions
    AppendToUserSubscriptions(subscriptionID)
    // push to back of DueQueue for this service
    PushToBackOfDueQueue(msg.Collector, msg.ServiceID, subscriptionID)
}
```

```go
// simply deletes mapping subscriptionID => subscription and removes ID from user list
// will not mutate duequeue
func HandleUnsubscribeMsg(ctx sdk.Context, msg UnsubscribeMsg) {
    DeleteSubscription(msg.Subscriber, msg.ServiceID)
}
```

```go
func HandleCollectMsg(ctx sdk.Context, msg CollectMsg) {
    if !SubscriptionExists(msg.ServiceID) {
        return NoSuchSubscriptionErr
    }
    Terms := GetTerms(ctx, msg.ServiceID)
    if msg.Collector != Terms.Collector {
        return InvalidCollectorErr
    }
    limit := msg.Limit
    // Only process a limited number of subscriptions defined in CollectMsg
    // This is to prevent msg handler from panicing with OutOfGasException if DueQueue gets too large
    for limit > 0 {
        // Each subscription period maintains its own queue
        // These queues are processed in a round robin fashion until all due subscriptions are collected
        // or msg.Limit has been reached
        for _, period := Terms.Period {
            if limit == 0 {
                // limit reached.
                // emit event indicating that not all due subscriptions have been collected
                return nil
            }
            queue := GetDueQueue(msg.Collector, msg.ServiceID, period)
            // top of the queue is "due"
            subscription := GetSubscription(queue[0])
            if subscription == nil {
                limit--
                continue
            }
            if subscription.LastPaid + period >= ctx.BlockTime {
                subscriptionID := PopOffFront(queue)
                if subscription.Limit == 0 {
                    DeleteSubscription(subscription.Subscriber, subscription.Name)
                    limit--
                    continue
                }
                err := bank.Send(subscription.Subscriber, Terms.Collector, Terms.Amount)
                if err != nil {
                    DeleteSubscription(subscription.Subscriber, subscription.Name)
                    limit--
                    continue
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
                // emit event indicating all due subscriptions reached with time for next duedate
                return nil
            }
        }
    }
    // limit reached.
    // emit event indicating that not all due subscriptions have been collected
    return nil
}
```


## Consequences

### Positive

- Adds a useful module to core-sdk that can be used by many future applications

### Negative

- One more module in `x/` folder

## References

- [Issue \#4642](https://github.com/cosmos/cosmos-sdk/issues/4642)

