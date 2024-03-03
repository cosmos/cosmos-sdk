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

This changes the current design in which each module that executes logic in begin and/or end block to the epoch module having a map of calls to make each interval. The epoch module will be responsible for calling the logic and/or passing messages to the respective modules. This does not replace the concept of beginblock and endblock but rather allows for modules to register their logic to be executed at different intervals.

This simplifies the need for their to be ordering in begin/endblock because the new design is around a caller to represent ordering. For example in the current design mint must come before distribution in begin block. With the proposed design mint and distribution would not have anything to register in beginblock and instead mint will only register a call to be made at a specific interval. Mint will then call distribution. 

### Module Registration

With the advent of propser based timestamps we can allow for modules to register logic to be executed based on a time. This will allow for more complex systems to be built.

EpochInfo is taken from the osmosis epochs module. The data structure defines what is needed for an epoch. 

```proto
// EpochInfo is a struct that describes the data going into
// a timer defined by the x/epochs module.
message EpochInfo {
  // identifier is a unique reference to this particular timer.
  string identifier = 1;
  // start_time is the time at which the timer first ever ticks.
  // If start_time is in the future, the epoch will not begin until the start
  // time.
  google.protobuf.Timestamp start_time = 2;
  // duration is the time in between epoch ticks.
  // In order for intended behavior to be met, duration should
  // be greater than the chains expected block time.
  // Duration must be non-zero.
  google.protobuf.Duration duration = 3;
  // current_epoch is the current epoch number, or in other words,
  // how many times has the timer 'ticked'.
  // The first tick (current_epoch=1) is defined as
  // the first block whose blocktime is greater than the EpochInfo start_time.
  int64 current_epoch = 4;
  // current_epoch_start_time describes the start time of the current timer
  // interval. The interval is (current_epoch_start_time,
  // current_epoch_start_time + duration] When the timer ticks, this is set to
  // current_epoch_start_time = last_epoch_start_time + duration only one timer
  // tick for a given identifier can occur per block.
  //
  // NOTE! The current_epoch_start_time may diverge significantly from the
  // wall-clock time the epoch began at. Wall-clock time of epoch start may be
  // >> current_epoch_start_time. Suppose current_epoch_start_time = 10,
  // duration = 5. Suppose the chain goes offline at t=14, and comes back online
  // at t=30, and produces blocks at every successive time. (t=31, 32, etc.)
  // * The t=30 block will start the epoch for (10, 15]
  // * The t=31 block will start the epoch for (15, 20]
  // * The t=32 block will start the epoch for (20, 25]
  // * The t=33 block will start the epoch for (25, 30]
  // * The t=34 block will start the epoch for (30, 35]
  // * The **t=36** block will start the epoch for (35, 40]
  google.protobuf.Timestamp current_epoch_start_time = 5;
  // epoch_counting_started is a boolean, that indicates whether this
  // epoch timer has began yet.
  bool epoch_counting_started = 6;
  // current_epoch_start_height is the block height at which the current epoch
  // started. (The block height at which the timer last ticked)
  int64 current_epoch_start_height = 7;
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

* Modules that require that use the epoch module will need to be modified to avoid using the epoch module, if an application developers does not want to use it. 

### Neutral

## Further Discussions

## Test Cases [optional]


## References

* [Osmosis Epoch](https://docs.osmosis.zone/osmosis-core/modules/epochs/)
