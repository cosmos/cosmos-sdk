Stake
=======

Proof of Stake related implementation including bonding and delegation transactions, inflation, fees, unbonding, etc.

Types:

**Candidate**

- **Status** (``CandidateStatus``) - Bonded status
- **Address** (``sdk.Address``) - Sender of BondTx - UnbondTx returns here
- **PubKey** (``crypto.PubKey``) - Pubkey of candidate
- **Assets** (``sdk.Rat``) - Total shares of a global hold pools
- **Liabilities** (``sdk.Rat``) - Total shares issued to a candidate's delegators
- **Description** (``Description``) - Description terms for the candidate
- **ValidatorBondHeight** (``int64``) - Earliest height as a bonded validator
- **ValidatorBondCounter** (``int64``) - Block-local tx index of validator change

**Description**

- **Moniker** (``string``) - Candidate's moniker/username
- **Identity** (``string``) - Real name
- **Website** (``string``) - URL of validator candidate's website
- **Details** (``string``) - Aditional optional details

**Validator**

- **Address** (``sdk.Address``) - Validator's address
- **PubKey** (``crypto.PubKey``) - Pubkey of the validator
- **Power** (``sdk.Rat``) - Total validator power held
- **Height** (``int64``) - Earliest block height as a validator
- **Counter** (``int16``) - Block-local tx index for resolving equal voting power & height

**DelegatorBond**

- **DelegatorAddr** (``sdk.Address``) - Delegator's address
- **CandidateAddr** (``sdk.Address``) - Validator candidate's address
- **Shares** (``sdk.Rat``) - Total shares delegated to the validator
- **Height** (``int64``) - Last height bond updated


Methods
^^^^^^^

``NewCandidate(address sdk.Address, pubKey crypto.PubKey, description Description)``

  Returns a new initialized ``Candidate``.

``NewDescription(moniker, identity, website, details string)``

  Returns the ``Description`` fields of the ``Candidate``.

``candidate.delegatorShareExRate()``

  Returns the exchange rate of global pool shares over delegator shares.

``candidate.validator()``

  Returns a copy of the ``Candidate`` as a ``Validator``.
  **Note**: Should only be called when the Candidate qualifies as a validator.

``validator.abciValidator(cdc *wire.Codec)``

  Returns an ``abci.Validator`` from stake validator type.

``validator.abciValidatorZero(cdc *wire.Codec)``

  Same as above, but its ``Power`` is set to 0 for validator updates.


Pool

**Pool**

- **TotalSupply** (``int64``) - Total supply of all tokens
- **BondedShares** (``sdk.Rat``) - Sum of all shares distributed for the Bonded Pool
- **UnbondedShares** (``sdk.Rat``) - Sum of all shares distributed for the Unbonded Pool
- **BondedPool** (``int64``) - Reserve of bonded tokens
- **UnbondedPool** (``int64``) - Reserve of unbonded tokens held with candidates
- **InflationLastTime** (``int64``) - Block which the last inflation was processed
- **Inflation** (``sdk.Rat``) - Current annual inflation rate

``pool.bondedRatio()``

  Get the bond ratio (``sdk.Rat``) of the global state .

``pool.bondedShareExRate()``

  Get the exchange rate of bonded token per issued share.

``pool.unbondedShareExRate()``

  Get the exchange rate of unbonded tokens held in candidates per issued share.

``pool.bondedToUnbondedPool(candidate Candidate)``

  Move a candidates asset pool from bonded to unbonded pool.

``pool.bondedToUnbondedPool(candidate Candidate)``

  Move a candidate's asset pool from unbonded to bonded pool.

``pool.candidateAddTokens(candidate Candidate, amount int64)``

  Add tokens to a candidate

``pool.candidateRemoveShares(candidate Candidate, shares sdk.Rat)``

  Remove shares from a candidate


Messages

**MsgDeclareCandidacy**

- **Description** (``Description``) -
- **CandidateAddr** (``sdk.Address``) -
- **PubKey** (``crypto.PubKey``) -
- **Bond** (``sdk.Coin``) -

**MsgEditCandidacy**

- **Description** (``Description``) -
- **CandidateAddr** (``sdk.Address``) -

**MsgDelegate**

- **DelegatorAddr** (``sdk.Address``) -
- **CandidateAddr** (``sdk.Address``) -
- **Bond** (``sdk.Coin``) -

**MsgUnbond**

- **DelegatorAddr** (``sdk.Address``) -
- **CandidateAddr** (``sdk.Address``) -
- **Shares** (``string``) -

``NewMsgDeclareCandidacy(candidateAddr sdk.Address, pubkey crypto.PubKey, bond sdk.Coin, description Description)``

  Returns a new ``MsgDeclareCandidacy`` message.

``NewMsgEditCandidacy(candidateAddr sdk.Address, description Description)``

  Returns a new ``MsgEditCandidacy`` message.

``NewMsgDelegate(delegatorAddr, candidateAddr sdk.Address, bond sdk.Coin)``

  Returns a new ``MsgDelegate`` instance.

``NewMsgUnbond(delegatorAddr, candidateAddr sdk.Address, shares string)``

  Returns a new ``MsgUnbond`` struct.
