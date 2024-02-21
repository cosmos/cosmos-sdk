# ADR 071: Epoch Module

## Changelog

* 19-Feb-2024: Initial Draft

## Status

Proposed

## Abstract

Currently in the Cosmos SDK all modules have the option to execute logic at the start and end of every block. If a module would like to do something at different intervals, but this logic is not exposed to clients. This makes it confusing when to expect things to be done and not. Secondly, if there are many modules that would like to execute logic at intervals coordinating which modules should run at which block can not be done. This can lead to countless modules running at the same block potentially causing issues with the block time.

## Context

Creating a coordination module that will allow modules to register their logic to be executed at different intervals. This will allow for better coordination of modules and allow for better understanding of when things will be executed. This will also allow for better control of the block time and allow for better performance.

The module does not handle storage of logic to be executed at a certain time, instead it is meant to be as stateless as possible. The implemention of module logic should not queue logic as well, it should execute but not return the data to respective location. For example, with staking val set updates should happen during the state transition but the returned value to comet for validator set updates only gets returned when the epoch is called. If the epoch is called at every block it will be the same as today. 

## Alternatives

Current design is kept and epoch logic is housed within specific modules instead of having the epoch module coordinate execution.

## Decision

A new epoch module is created that will allow for modules to register their logic to be executed at different intervals. Modules can register specific calls to be called at specified intervals. The epoch module will be responsible for executing the logic at the specified intervals.

This changes the current design in which each module that executes logic in begin and/or end block to the epoch module having a map of calls to make each interval. The epoch module will be responsible for calling the logic and/or passing messages to the respective modules. 

This simplifies the need for their to be ordering in begin/endblock because the new design is around a caller to represent ordering. For example in the current design mint must come before distribution in begin block. With the proposed design mint and distribution would not have anything to register in beginblock and instead mint will only register a call to be made at a specific interval. Mint will then call distribution. 

### Module Registration

With the advent of propser based timestamps we can allow for modules to register logic to be executed based on a time. This will allow for more complex systems to be built.

```proto
message EpochInfo{

}
```

```go 
type Module interface {
    RegisterBeginEpochCalls() map[time.seconds]func() error
    RegisterEndEpochCalls() map[time.seconds]func() error
}
```

A module will write a functions that returns an error to execute the specific logic at an interval. In the future with internal message passing this can be done with a message instead of a function. If a module does want to use a message they can do this as well and avoid the need to write the function on the keeper. 

## Consequences

* The new modules becomes required for all modules that would allow users to register logic to be executed at different intervals.

### Positive

* Modules can register logic to be executed at different intervals.
* Intervals for execution can be queried by clients

### Negative

* Module requirements are increased. If a user wants to execute something every block and avoid the epoch module they will need to still use the epoch module. 

### Neutral

> {neutral consequences}

## Further Discussions

> While an ADR is in the DRAFT or PROPOSED stage, this section should contain a
> summary of issues to be solved in future iterations (usually referencing comments
> from a pull-request discussion).
> 
> Later, this section can optionally list ideas or improvements the author or
> reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus
changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
