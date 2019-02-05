TODO: Ensure this is up to date.

# Stores

This document provides a bit more insight as to the purpose of several related
prefixed areas of the staking store which are accessed in `x/staking/keeper.go`.

# IAVL Store 

## Validators
 - Prefix Key Space:    ValidatorsKey
 - Key/Sort:            Validator operator address (`valoper`)
 - Value:               Validator object
 - Contains:            All validator records independent of being bonded or not
 - Used For:            Retrieve validator from operator address, general validator retrieval

## Validators By Power
 - Prefix Key Space:    ValidatorsByPowerKey
 - Key/Sort:            Validator power (equivalent bonded shares) then block
                        height then transaction order
 - Value:               Validator operator address (`valoper`)
 - Contains:            All validator records independent of being bonded or not
 - Used For:            Determining which validators ought to be bonded, as sorted by power

## Validators Bonded
 - Prefix Key Space:    ValidatorsBondedKey
 - Key/Sort:            Validator pubkey address (`valconspub`, NOTE: same as Tendermint)
 - Value:               Validator operator address
 - Contains:            Only currently bonded validators
 - Used For:            Retrieving the list of all currently bonded validators. When updating
                        for a new validator entering the validator set, we may want to loop
                        through this set to determine who we've kicked out.
                        Also used for retrieving validator by Tendermint index.

# Transient Store

The transient store persists between transations but not between blocks.

TODO: Finish me, what do we store in the tranient store?
