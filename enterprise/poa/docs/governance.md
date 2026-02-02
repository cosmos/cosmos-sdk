# Governance

## Overview

The PoA module integrates with Cosmos SDK governance to restrict participation to authorized validators only. Unlike standard governance that uses bonded tokens for voting weight, PoA governance uses validator power as the basis for voting.

## Validator-Only Governance

The PoA module restricts governance participation to authorized validators only through governance hooks.

**Governance Hooks** ([`x/poa/keeper/hooks.go`](../x/poa/keeper/hooks.go))

The module implements `govtypes.GovHooks`:

1. **AfterProposalSubmission**: Only authorized validators can submit proposals
2. **AfterProposalDeposit**: Only authorized validators can deposit on proposals
3. **AfterProposalVote**: Only authorized validators can vote

**Authorized Validator Definition**:
- Registered in PoA module
- Power > 0
- Has valid operator address

**Rejected Actions**:
- If non-validator attempts governance action → transaction fails
- If validator has power = 0 → transaction fails
- Error: "voter X is not an active PoA validator"

**Location**: [`x/poa/keeper/governance.go:92`](../x/poa/keeper/governance.go#L92)

## Voting Power

**Custom Vote Tallying** ([`x/poa/keeper/governance.go:18`](../x/poa/keeper/governance.go#L18)). An example of the wiring can be found in the [SimApp](../simapp/app.go#L197-214).

Standard governance uses staked tokens as voting weight. PoA governance uses validator power:

1. **Vote Collection**: System iterates all votes on a proposal
2. **Validator Check**: For each vote, verify voter is active PoA validator
3. **Weight Calculation**: Use validator's power as voting weight
4. **Weighted Options**: Supports split votes, exactly like x/staking in traditional POS governance (e.g., 70% Yes, 30% Abstain)
5. **Tally Results**: Sum weighted votes by option

### Vote Tallying Algorithm

**Voting Power Formula**:

$$
V_i = P_i
$$

Where:
- $V_i$ = voting power of validator $i$
- $P_i$ = validator power (consensus weight)

**Weighted Vote Calculation**:

For a validator casting a split vote across multiple options:

$$
W_{i,o} = V_i \times w_{i,o}
$$

Where:
- $W_{i,o}$ = vote weight from validator $i$ for option $o$
- $w_{i,o}$ = weight assigned to option $o$ by validator $i$ (where $\sum_{o} w_{i,o} = 1$)

**Total Tally per Option**:

$$
T_o = \sum_{i \in \text{voters}} W_{i,o}
$$

Where:
- $T_o$ = total votes for option $o$
- Sum over all validators who voted

### Example

**Validator A**: $P_A = 100$, votes 100% Yes
- $W_{A,\text{Yes}} = 100 \times 1.0 = 100$

**Validator B**: $P_B = 50$, votes 60% Yes, 40% No
- $W_{B,\text{Yes}} = 50 \times 0.6 = 30$
- $W_{B,\text{No}} = 50 \times 0.4 = 20$

**Totals**:
- $T_{\text{Yes}} = 130$
- $T_{\text{No}} = 20$

## Proposal Lifecycle

### 1. Proposal Submission

**[MsgSubmitProposal](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/gov/v1/tx.proto#L54-L65)** (standard x/gov module)

When a proposal is submitted:

1. Standard governance validates the proposal content
2. `AfterProposalSubmission` hook is called
3. PoA module checks if proposer is authorized validator:
   - Look up proposer by operator address
   - Verify validator exists and has $P > 0$
   - If not active, reject with error
4. If valid, proposal enters deposit period

**Restriction**: Only authorized validators can submit proposals, preventing spam from non-consensus participants.

### 2. Deposit Period

**[MsgDeposit](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/gov/v1/tx.proto#L90-L98)** (standard x/gov module)

When a deposit is made:

1. Standard governance processes the deposit
2. `AfterProposalDeposit` hook is called
3. PoA module checks if depositor is authorized validator
4. If deposit threshold reached, proposal moves to voting period

**Restriction**: Only authorized validators can deposit, ensuring only consensus participants can advance proposals.

### 3. Voting Period

**[MsgVote](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/gov/v1/tx.proto#L100-L108)** or **[MsgVoteWeighted](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/gov/v1/tx.proto#L110-L118)** (standard x/gov module)

When a vote is cast:

1. Standard governance records the vote
2. `AfterProposalVote` hook is called
3. PoA module validates voter is authorized validator
4. If invalid, transaction fails

**Vote Options**:
- `Yes`: Support the proposal
- `No`: Oppose the proposal
- `NoWithVeto`: Oppose and veto (can burn deposits if threshold met)
- `Abstain`: Participate in quorum without taking a position

**Weighted Voting**: Validators can split their vote across multiple options, with weights summing to 1.

### 4. Vote Tallying

At the end of the voting period, the [custom tally function](#vote-tallying-algorithm) is called:

**NewPoACalculateVoteResultsAndVotingPowerFn** ([`x/poa/keeper/governance.go:18`](../x/poa/keeper/governance.go#L18))

1. Iterate all votes on the proposal
2. For each vote, look up the validator by voter address
3. If validator is not active ($P \leq 0$), skip the vote
4. Otherwise, use validator power as voting weight
5. For weighted votes, distribute power across options
6. Sum all weighted votes by option
7. Apply standard governance thresholds:
   - Quorum: Minimum participation percentage
   - Threshold: Minimum "Yes" percentage to pass
   - Veto: Maximum "NoWithVeto" percentage before rejection

**Result**: Proposal passes, fails, or is vetoed based on power-weighted votes.

## Implementation Details

### Governance Hooks

**Location**: [`x/poa/keeper/hooks.go`](../x/poa/keeper/hooks.go)

The module implements the `govtypes.GovHooks` interface:

```
type GovHooks interface {
    AfterProposalSubmission(ctx, proposalID, depositorAddr)
    AfterProposalDeposit(ctx, proposalID, depositorAddr)
    AfterProposalVote(ctx, proposalID, voterAddr)
    // ... other hooks
}
```

Each hook implementation:
1. Extracts the operator address from the context
2. Looks up the validator by operator address
3. Checks if validator exists and has power > 0
4. Returns error if validation fails

### Custom Tally Function

**Location**: [`x/poa/keeper/governance.go:18`](../x/poa/keeper/governance.go#L18)

The tally function replaces the standard governance tally:

```
func NewPoACalculateVoteResultsAndVotingPowerFn(keeper) TallyFn {
    return func(ctx, proposal) (totalVotingPower, results) {
        // Iterate votes
        for vote in votes(proposal) {
            validator = keeper.GetValidatorByOperator(vote.voter)
            if validator == nil || validator.Power <= 0 {
                continue // Skip non-authorized validators
            }

            // Add validator power to total
            totalVotingPower += validator.Power

            // Apply vote weights
            for option, weight in vote.options {
                results[option] += validator.Power * weight
            }
        }
        return totalVotingPower, results
    }
}
```

## Governance Parameters

The standard governance module parameters still apply:

- **MinDeposit**: Minimum tokens required to enter voting period
- **MaxDepositPeriod**: Time limit for reaching minimum deposit
- **VotingPeriod**: Duration of the voting period
- **Quorum**: Minimum participation rate (fraction of total power)
- **Threshold**: Minimum "Yes" rate to pass (fraction of non-abstain votes)
- **VetoThreshold**: Maximum "NoWithVeto" rate before rejection

**Key Difference**: Quorum is calculated as a percentage of total validator power, not total bonded tokens.

## Security Considerations

1. **Validator Exclusivity**:
   - Only authorized validators (power > 0) can participate
   - Prevents sybil attacks through unauthorized validator spam
   - Ensures governance represents actual consensus participants

2. **Power-Based Voting**:
   - Voting weight tied to consensus power
   - Admin controls power distribution, thus controls governance indirectly

3. **Admin Governance Control**:
   - Admin can change validator power at any time
   - Admin can effectively control governance by adjusting power
   - Consider multi-sig admin or governance-controlled admin changes

4. **Proposal Spam Prevention**:
   - Restricting submissions to authorized validators reduces spam
   - Deposit requirements still apply
   - Validators have reputational stake in proposal quality

## Comparison to Standard Governance

| Aspect | Standard Cosmos Governance | PoA Governance |
|--------|---------------------------|----------------|
| Who can vote | Token holders (delegators + validators) | Authorized validators only |
| Voting weight | Bonded tokens | Validator power |
| Who can propose | Anyone with min deposit | Authorized validators only |
| Who can deposit | Anyone | Authorized validators only |
| Vote tallying | Sum of bonded tokens | Sum of validator power |
| Quorum calculation | % of bonded tokens | % of total validator power |
| Admin control | No direct control | Admin controls power → controls votes |

## Example Governance Flow

**Scenario**: Validator A wants to propose a parameter change

1. **Submit Proposal**:
   - Validator A (power = 40) submits `MsgSubmitProposal`
   - Hook verifies A is authorized validator
   - Proposal enters deposit period

2. **Reach Deposit**:
   - Validator B (power = 30) deposits
   - Validator C (power = 30) deposits
   - Deposit threshold reached → voting period starts

3. **Voting**:
   - Validator A: 100% Yes (40 power → 40 Yes votes)
   - Validator B: 60% Yes, 40% No (30 power → 18 Yes, 12 No)
   - Validator C: 100% Abstain (30 power → 30 Abstain)
   - Total power: 100 (all authorized validators)

4. **Tally**:
   - Total voting power: 100 (all voted)
   - Quorum: 100/100 = 100% ✓ (assuming 33% quorum)
   - Results: 58 Yes, 12 No, 30 Abstain (out of 70 non-abstain)
   - Threshold: 58/70 = 82.9% Yes ✓ (assuming 50% threshold)
   - **Proposal passes**
