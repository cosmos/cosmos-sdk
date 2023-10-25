# Pruning

## Overview

Pruning is the mechanism for deleting old versions data from both state storage and commitment. The pruning operation is done in background and is triggered periodically by the root store.

## Pruning Options

Generally, there are two configurable parameters for pruning options:

- `pruning-keep-recent`: the number of recent versions to keep.
- `pruning-interval`: the interval between two pruning operations.

Different options will be applied to the state storage and commitment. The pruning option have an effect on the snapshot operation, but it will not manage the conflict resolution in SDK, it is the responsibility of the dedicated backend.
