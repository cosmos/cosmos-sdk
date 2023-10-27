# Pruning

## Overview

Pruning is the mechanism for deleting old versions data from both state storage and commitment. The pruning operation is triggered periodically.

## Pruning Options

Generally, there are three configurable parameters for pruning options:

- `pruning-keep-recent`: the number of recent versions to keep.
- `pruning-interval`: the interval between two pruning operations.
- `pruning-sync`: the flag to sync/async the pruning operation.

Different options will be applied to the state storage and commitment. The pruning option have an effect on the snapshot operation, but it will not manage the conflict resolution in SDK, it is the responsibility of the dedicated backend.
