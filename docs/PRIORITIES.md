## Fees
- Collection
  - Gas price based on parameter
  - (which gets changed automatically)
  - https://github.com/cosmos/cosmos-sdk/issues/1921
  - Per block gas usage as %
  - Windowing function
    - Block N,
    - For Block N-x ~ N, get average of %
  - Should take into account time.
  - Standard for querying this price // needs to be used by UX.
- Distribution
  - MVP: 1 week, 1 week for testing.

## Governance v2
- V1 is just text proposals
  - Software upgrade stuff
    - https://github.com/cosmos/cosmos-sdk/issues/1734#issuecomment-407254938
    - https://github.com/cosmos/cosmos-sdk/issues/1079
- We need to test auto-flipping w/ threshold voting power.
- Super simple:
  - Only use text proposals
  - On-chain mechanism for agreeing on when to "flip" to new functionality

## Staking/Slashing/Stability
- Unbonding state for validators https://github.com/cosmos/cosmos-sdk/issues/1676
- current: downtime, double signing during unbonding
- who gets slashed when -- needs review about edge cases
- need to communicate to everyone that lite has this edge case
	- Issues:
    - https://github.com/cosmos/cosmos-sdk/issues/1378
    - https://github.com/cosmos/cosmos-sdk/issues/1348
    - https://github.com/cosmos/cosmos-sdk/issues/1440
  * Est Difficulty: Hard
  * _*Note:*_ This feature needs to be fully fleshed out. Will require a meeting between @jaekwon, @cwgoes, @rigel, @zaki, @bucky to discuss the issues. @jackzampolin to facilitate.

## Vesting
- 24 accounts with NLocktime
- “No funds can be transferred before timelock”
- New atoms and such can be withdrawn right way
- Requires being able to send fees and inflation to new account

## Multisig
- Make it work with Cli
- ADR

## Reserve Pool
- No withdrawing from it at launch

## Other:
- Need to update for NextValidatorSet - need to upgrade SDK for it
- Need to update for new ABCI changes - error string, tags are list of lists, proposer in header
- Inflation ? 

## Gas
- Calculate gas

## Reward proposer
- Requires tendermint changes

# Lower priority

## Circuit Breaker
- Kinda needed for enabling txs.

## Governance proposal changes
- V2 is parameter changes

## Slashing/Stability
- tendermint evidence: we don’t yet slash byzantine signatures (signing at all) when not bonded.

# Other priority

## gaiad // gaiacli
- Documentation // language

## gaialite
- Documentation // language
