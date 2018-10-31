# Stores

This document provided a bit more insight as to the purpose of several related
prefixed areas of the staking store which are accessed in `x/stake/keeper.go`.

# IAVL Store 

## Validators
 - Prefix Key Space:    ValidatorsKey
 - Key/Sort:            Validator Operator Address
 - Value:               Validator Object
 - Contains:            All Validator records independent of being bonded or not
 - Used For:            Retrieve validator from operator address, general validator retrieval

## Validators By Power
 - Prefix Key Space:    ValidatorsByPowerKey
 - Key/Sort:            Validator Power (equivalent bonded shares) then Block
                        Height then Transaction Order
 - Value:               Validator Operator Address
 - Contains:            All Validator records independent of being bonded or not
 - Used For:            Determining who the top validators are whom should be bonded

## Validators Cliff Power
 - Prefix Key Space:    ValidatorCliffKey
 - Key/Sort:            single-record
 - Value:               Validator Power Key (as above store)
 - Contains:            The cliff validator (ex. 100th validator) power
 - Used For:            Efficient updates to validator status

## Validators Bonded
 - Prefix Key Space:    ValidatorsBondedKey
 - Key/Sort:            Validator PubKey Address (NOTE same as Tendermint)
 - Value:               Validator Operator Address
 - Contains:            Only currently bonded Validators
 - Used For:            Retrieving the list of all currently bonded validators when updating
                        for a new validator entering the validator set we may want to loop
                        through this set to determine who we've kicked out.
                        retrieving validator by tendermint index

# Transient Store 

The transient store persists between transations but not between blocks 

