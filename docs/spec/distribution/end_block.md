# End Block

At each endblock, the fees received are sorted to the proposer, community fund,
and global pool.  When the validator is the proposer of the round, that
validator (and their delegators) receives between 1% and 5% of fee rewards, the
reserve tax is then charged, then the remainder is distributed socially by
voting power to all validators including the proposer validator.  The amount of
proposer reward is calculated from pre-commits Tendermint messages. All
provision rewards are added to a provision reward pool which validator holds
individually (`ValidatorDistribution.ProvisionsRewardPool`). 

```
func SortFees(feesCollected sdk.Coins, global Global, proposer ValidatorDistribution, 
              sumPowerPrecommitValidators, totalBondedTokens, communityTax sdk.Dec)

     feesCollectedDec = MakeDecCoins(feesCollected)
     proposerReward = feesCollectedDec * (0.01 + 0.04 
                       * sumPowerPrecommitValidators / totalBondedTokens)
     proposer.ProposerPool += proposerReward
     
     communityFunding = feesCollectedDec * communityTax
     global.CommunityFund += communityFunding
     
     poolReceived = feesCollectedDec - proposerReward - communityFunding
     global.Pool += poolReceived
     global.EverReceivedPool += poolReceived
     global.LastReceivedPool = poolReceived

     SetValidatorDistribution(proposer)
     SetGlobal(global)
```
