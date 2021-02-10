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

# The Share Abstraction

At any given point in time, each validator has a number of tokens, `T`, and has a number of shares issued, `S`.
Each delegator, `i`, holds a number of shares, `S_i`.
The number of tokens is the sum of all tokens delegated to the validator, plus the rewards, minus the slashes.

The delegator is entitled to a portion of the underlying tokens proportional to their proportion of shares.
So delegator `i` is entitled to `T * S_i / S` of the validator's tokens.

When a delegator delegates new tokens to the validator, they receive a number of shares proportional to their contribution.
So when delegator `j` delegates `T_j` tokens, they receive `S_j = S * T_j / T` shares.
The total number of tokens is now `T + T_j`, and the total number of shares is `S + S_j`.
`j`s proportion of the shares is the same as their proportion of the total tokens contributed: `(S + S_j) / S = (T + T_j) / T`.

A special case is the initial delegation, when `T = 0` and `S = 0`, so `T_j / T` is undefined.
For the initial delegation, delegator `j` who delegates `T_j` tokens receive `S_j = T_j` shares.
So a validator that hasn't received any rewards and has not been slashed will have `T = S`.