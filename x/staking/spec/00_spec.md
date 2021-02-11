# Validator tokens and Consensus Power

Total consensus power of a delegator can be calculated from the number of validator tokens they hold, and vice versa. 
The equation is as follows:

```
tokens = consensus_power * power_reduction
```

`power_reduction` is a constant parameter set to `10**6` by default.

TODO: Is consensus power a network-wide measure or per validator?

# Staker vs Delegator vs Validator
The Staker and Delegator are actual EOAs.
The Validator is an internal concept (with an address), controlled by the staker.

