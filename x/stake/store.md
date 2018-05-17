# Stores

This document provided a bit more insight as to the purpose of several related
prefixed areas of the staking store which are accessed in `x/stake/keeper.go`.


## Validators 
 - Prefix Key Space:    ValidatorsKey
 - Key/Sort:            Validator Owner Address
 - Contains:            All Validator records independent of being bonded or not
 - Used For:            Retrieve validator from owner address, general validator retrieval 

## Validators By Power
 - Prefix Key Space:    ValidatorsByPowerKey
 - Key/Sort:            Validator Power (equivalent bonded shares) then Block
                        Height then Transaction Order
 - Contains:            All Validator records independent of being bonded or not
 - Used For:            Determining who the top validators are whom should be bonded

## Validators Bonded 
 - Prefix Key Space:    ValidatorsBondedKey
 - Key/Sort:            Validator PubKey Address (NOTE same as Tendermint)
 - Contains:            Only currently bonded Validators
 - Used For:            Retrieving the list of all currently bonded validators when updating 
                        for a new validator entering the validator set we may want to loop 
                        through this set to determine who we've kicked out.
                        retrieving validator by tendermint index

## Tendermint Updates
 - Prefix Key Space:    TendermintUpdatesKey
 - Key/Sort:            Validator Owner Address
 - Contains:            Validators are queued to affect the consensus validation set in Tendermint
 - Used For:            Informing Tendermint of the validator set updates, is used only intra-block, as the
                        updates are applied then cleared on endblock
