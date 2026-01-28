# Fee Distribution

## Overview

The PoA module implements a custom fee distribution mechanism based on validator power. Unlike the standard Cosmos SDK x/distribution module, PoA uses a checkpoint-based system to allocate fees proportionally to validators without automatic distribution.

## How Fees Accumulate

Fees flow through the PoA system differently than standard Cosmos SDK:

1. **Block Fees**: Transaction fees collected in each block go to the standard `fee_collector` module account

2. **Checkpoint System**: Allocated fees are updated for validators when:
   - Any validator power changes
   - Any validator withdraws fees

**Why Checkpointing?**: Ensures fair distribution when power changes. If power changes mid-period, fees are allocated based on old power distribution before the change takes effect.

**Location**: [`x/poa/keeper/distribution.go:18`](../x/poa/keeper/distribution.go#L18)

## Distribution Algorithm

### Checkpoint-Based Allocation

The PoA module uses a checkpoint system to allocate fees fairly when validator power changes. Rather than distributing fees actively at every block, allocation efficiently happens at discrete checkpoints.

**Checkpoint Triggers**:
- Any validator power change (via `MsgUpdateValidators`)
- Any fee withdrawal (via `MsgWithdrawFees`)

**Unallocated Fees Calculation**:

At checkpoint time $t$, calculate unallocated fees:

$$
U_t = B_{\text{collector}}(t) - A_{\text{total}}(t)
$$

Where:
- $U_t$ = unallocated fees at checkpoint $t$
- $B_{\text{collector}}(t)$ = current balance in the fee collector module account
- $A_{\text{total}}(t) = \sum_{i=1}^{n} F_i(t)$ = sum of all previously- allocated fees across all validators (0 if no checkpoints have been done)

**Proportional Share Allocation**:

For each active validator $i$ (where $P_i(t) > 0$), allocate a share proportional to their power:

$$
S_i(t) = U_t \times \frac{P_i(t)}{P_{\text{total}}(t)}
$$

Where:
- $S_i(t)$ = share allocated to validator $i$ at checkpoint $t$
- $P_i(t)$ = voting power of validator $i$ at checkpoint $t$
- $P_{\text{total}}(t) = \sum_{j=1}^{n} P_j(t)$ = sum of all validator powers

**Accumulated Fees Update**:

After allocation, update each validator's accumulated fees:

$$
F_i(t+1) = F_i(t) + S_i(t)
$$

Where:
- $F_i(t)$ = validator $i$'s accumulated fees before checkpoint
- $F_i(t+1)$ = validator $i$'s accumulated fees after checkpoint
- $S_i(t)$ = share allocated in this checkpoint

**Total Allocated Tracking**:

Update the global allocated tracker:

$$
A_{\text{total}}(t+1) = A_{\text{total}}(t) + U_t
$$

After this checkpoint, $A_{\text{total}}(t+1) = B_{\text{collector}}(t)$ (all fees are now allocated).

### Example Checkpoint Sequence

**Initial State** (before checkpoint):
- Fee collector balance: $B_{\text{collector}} = 1000$ tokens
- Total allocated: $A_{\text{total}} = 400$ tokens (from previous checkpoints)
- Validator A: $P_A = 50$, $F_A = 200$ tokens allocated
- Validator B: $P_B = 50$, $F_B = 200$ tokens allocated
- Total power: $P_{\text{total}} = 100$

**Admin Action**: Admin submits `MsgUpdateValidators` to change power distribution to 30/70

**Checkpoint Triggered** (before power change takes effect):

1. Calculate unallocated: $U = 1000 - 400 = 600$ tokens

2. Allocate shares based on **current power** (50/50):
   - Validator A: $S_A = 600 \times \frac{50}{100} = 300$ tokens
   - Validator B: $S_B = 600 \times \frac{50}{100} = 300$ tokens

3. Update accumulated fees:
   - Validator A: $F_A = 200 + 300 = 500$ tokens
   - Validator B: $F_B = 200 + 300 = 500$ tokens

4. Update total allocated: $A_{\text{total}} = 400 + 600 = 1000$ tokens

**After Checkpoint** - Power Change Applied:
- Validator A: $P_A = 30$ (new power for future blocks)
- Validator B: $P_B = 70$ (new power for future blocks)
- All 1000 tokens now allocated ($A_{\text{total}} = B_{\text{collector}}$)
- Each validator has updated $F_i$ available for withdrawal

**Why This Matters**: Validator A earned 300 tokens (50% share) based on their power during the period when those fees were collected. After the checkpoint, their power drops to 30%, so future fees will be split 30/70. Checkpointing ensures validators are rewarded based on the work they actually performed.

**Precision**: Uses `DecCoins` (decimal coins) to prevent rounding dust accumulation. Each validator tracks fractional amounts that are too small to withdraw.

## Withdrawing Fees

**MsgWithdrawFees** ([`x/poa/keeper/msg_server.go:91`](../x/poa/keeper/msg_server.go#L91))

Any validator operator can withdraw accumulated fees:

1. **Submit Withdrawal**: Signed by operator address
2. **Checkpoint**: System checkpoints all validators first (allocates any pending fees)
3. **Truncate**: Decimal coins truncated to whole coins
4. **Transfer**: Coins transferred from `fee_collector` to operator address
5. **Update Tracking**: Total allocated decreases by withdrawn amount
6. **Remainder**: Decimal remainder stays in validator's allocated balance

**Example**:
```
Validator has: 100.7543 utokens allocated
Withdrawal:    100 utokens transferred to operator
Remainder:     0.7543 utokens remain allocated (less than least significant utoken digit)
```

**Location**: [`x/poa/keeper/distribution.go:106`](../x/poa/keeper/distribution.go#L106)

## Withdrawal Formula

When validator $i$ withdraws fees:

$$
W_i = \lfloor F_i \rfloor
$$

$$
F_i' = F_i - W_i
$$

$$
A_{\text{total}}' = A_{\text{total}} - W_i
$$

Where:
- $W_i$ = amount withdrawn (truncated to integer coins)
- $F_i$ = validator's allocated fees before withdrawal
- $F_i'$ = validator's allocated fees after withdrawal (decimal remainder)
- $\lfloor F_i \rfloor$ = floor function (truncate decimals)
- $A_{\text{total}}'$ = updated total allocated across all validators

## Security Considerations

1. **Decimal Precision**:
   - Uses DecCoins to prevent dust accumulation
   - Validators track fractional amounts
   - Remainders preserved across withdrawals
   - Prevents rounding errors from accumulating
