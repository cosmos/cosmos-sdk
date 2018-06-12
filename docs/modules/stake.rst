Stake
=====

Proof of Stake related implementation including bonding and delegation transactions, inflation, fees, unbonding, etc.

Types
-----

**Candidate**
^^^^^^^^^^^^^

- **Status** (``CandidateStatus``) - Bonded status
- **Address** (``sdk.Address``) - Sender of BondTx - UnbondTx returns here
- **PubKey** (``crypto.PubKey``) - Pubkey of candidate
- **Assets** (``sdk.Rat``) - Total shares of a global hold pools
- **Liabilities** (``sdk.Rat``) - Total shares issued to a candidate's delegators
- **Description** (``Description``) - Description terms for the candidate
- **ValidatorBondHeight** (``int64``) - Earliest height as a bonded validator
- **ValidatorBondCounter** (``int64``) - Block-local tx index of validator change


Candidate defines the total amount of bond shares and their exchange rate to
coins. Accumulation of interest is modelled as an in increase in the
exchange rate, and slashing as a decrease.  When coins are delegated to this
candidate, the candidate is credited with a DelegatorBond whose number of
bond shares is based on the amount of coins delegated divided by the current
exchange rate. Voting power can be calculated as total bonds multiplied by
exchange rate.

Methods
"""""""

``NewCandidate(address sdk.Address, pubKey crypto.PubKey, description Description)``
************************************************************************************

  Returns: ``Candidate``

  Creates a new validator candidate.

``candidate.delegatorShareExRate()``
************************************

  Returns: ``sdk.Rat``

  Returns the exchange rate of global pool shares over delegator shares.

``candidate.validator()``
*************************

  Returns: ``Validator``

  Returns a copy of the Candidate as a Validator.

  **Note**: Should only be called when the Candidate qualifies as a validator.


**Description**
^^^^^^^^^^^^^^^

- **Moniker** (``string``) - Candidate's moniker/username
- **Identity** (``string``) - Real name
- **Website** (``string``) - URL of validator candidate's website
- **Details** (``string``) - Aditional optional details

Methods
"""""""

``NewDescription(moniker, identity, website, details string)``
**************************************************************

  Returns: ``Description``

  Returns the description fields of the candidate.

**Validator**
^^^^^^^^^^^^^

- **Address** (``sdk.Address``) - Validator's address
- **PubKey** (``crypto.PubKey``) - Pubkey of the validator
- **Power** (``sdk.Rat``) - Total validator power held
- **Height** (``int64``) - Earliest block height as a validator
- **Counter** (``int16``) - Block-local tx index for resolving equal voting power & height

A ``Validator`` is one of the top ``Candidate`` s with most voting power.

Methods
"""""""

``validator.abciValidator(cdc *wire.Codec)``
********************************************

  Returns: ``abci.Validator``

  Returns an ABCI Validator from stake validator type.

``validator.abciValidatorZero(cdc *wire.Codec)``
************************************************

  Returns: ``abci.Validator``

  Same as above, but its ``Power`` is set to 0 for validator updates.


**DelegatorBond**
^^^^^^^^^^^^^^^^^

- **DelegatorAddr** (``sdk.Address``) - Delegator's address
- **CandidateAddr** (``sdk.Address``) - Validator candidate's address
- **Shares** (``sdk.Rat``) - Total shares delegated to the validator
- **Height** (``int64``) - Last height bond updated

DelegatorBond represents the bond with tokens held by an account.  It is
owned by one delegator, and is associated with the voting power of one
pubKey.

**Pool**
^^^^^^^^

- **TotalSupply** (``int64``) - Total supply of all tokens
- **BondedShares** (``sdk.Rat``) - Sum of all shares distributed for the Bonded Pool
- **UnbondedShares** (``sdk.Rat``) - Sum of all shares distributed for the Unbonded Pool
- **BondedPool** (``int64``) - Reserve of bonded tokens
- **UnbondedPool** (``int64``) - Reserve of unbonded tokens held with candidates
- **InflationLastTime** (``int64``) - Block which the last inflation was processed
- **Inflation** (``sdk.Rat``) - Current annual inflation rate

Methods
"""""""

``pool.equal(p2 Pool)``
***********************

  Returns: ``bool``

  Checks if a pool is equal to another.

``pool.bondedRatio()``
**********************

  Returns: ``sdk.Rat``

  Gets the bond ratio of the global state.

``pool.bondedShareExRate()``
****************************

  Returns: ``sdk.Rat``

  Gets the exchange rate of bonded token per issued share.

``pool.unbondedShareExRate()``
******************************

  Returns: ``sdk.Rat``

  Gets the exchange rate of unbonded tokens held in candidates per issued share.

``pool.bondedToUnbondedPool(candidate Candidate)``
**************************************************

  Returns: ``Pool``, ``Candidate``

  Move a candidates asset pool from bonded to unbonded pool.

``pool.bondedToUnbondedPool(candidate Candidate)``
**************************************************

  Returns: ``Pool``, ``Candidate``

  Moves a candidate's asset pool from unbonded to bonded pool.

``pool.candidateAddTokens(candidate Candidate, amount int64)``
**************************************************************

  Returns: ``Pool``, ``Candidate``, ``sdk.Rat``

  Adds tokens to a candidate.

``pool.candidateRemoveShares(candidate Candidate, shares sdk.Rat)``
*******************************************************************

  Returns: ``Pool``, ``Candidate``, ``int64``

  Removes shares from a candidate.


Messages
--------

**MsgDeclareCandidacy**
^^^^^^^^^^^^^^^^^^^^^^^

- **Description** (``Description``) -
- **CandidateAddr** (``sdk.Address``) -
- **PubKey** (``crypto.PubKey``) -
- **Bond** (``sdk.Coin``) -

Methods
"""""""

``NewMsgDeclareCandidacy(candidateAddr sdk.Address, pubkey crypto.PubKey, bond sdk.Coin, description Description)``
*******************************************************************************************************************

  Returns: ``MsgDeclareCandidacy``

  Creates a message to declare a candidate.

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.


**MsgEditCandidacy**
^^^^^^^^^^^^^^^^^^^^

- **Description** (``Description``) -
- **CandidateAddr** (``sdk.Address``) -

Methods
"""""""

``NewMsgEditCandidacy(candidateAddr sdk.Address, description Description)``
***************************************************************************

Returns: ``MsgEditCandidacy``

Creates a message to edit a candidate's info.

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.

**MsgDelegate**
^^^^^^^^^^^^^^^

- **DelegatorAddr** (``sdk.Address``) -
- **CandidateAddr** (``sdk.Address``) -
- **Bond** (``sdk.Coin``) -

Methods
"""""""

``NewMsgDelegate(delegatorAddr, candidateAddr sdk.Address, bond sdk.Coin)``
***************************************************************************

  Returns: ``NewMsgDelegate``

  Creates a new message to delegate bonds to a candidate.

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message.

**MsgUnbond**
^^^^^^^^^^^^^

- **DelegatorAddr** (``sdk.Address``) -
- **CandidateAddr** (``sdk.Address``) -
- **Shares** (``string``) -

Methods
"""""""

``NewMsgUnbond(delegatorAddr, candidateAddr sdk.Address, shares string)``
*************************************************************************

  Returns: ``MsgUnbond``

  Creates a new message to unbond shares.

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.

Handlers
--------

Staking Handlers
^^^^^^^^^^^^^^^^

Methods
"""""""

``NewHandler(k Keeper)``
************************

  Returns: ``sdk.Handler``

  Creates a new Handler according to the keeper message type. This handler can be
  one of the following: ``handleMsgDeclareCandidacy``, ``handleMsgEditCandidacy``,
  ``handleMsgDelegate`` or ``handleMsgUnbond``.


``NewEndBlocker(k Keeper)``
***************************

  Returns: ``sdk.EndBlocker``

  Generates an EndBlocker that performs tick functionality.

``InitGenesis(ctx sdk.Context, k Keeper, data GenesisState)``
*************************************************************

  Returns: ``nil``

  Stores genesis parameters

``WriteGenesis(ctx sdk.Context, k Keeper)``
*******************************************

  Returns: ``GenesisState``

  Creates an output genesis parameters (*i.e.* ``Pool``, ``Params``, ``Candidates`` and ``Bonds``)

``handleMsgDeclareCandidacy(ctx sdk.Context, msg MsgDeclareCandidacy, k Keeper)``
*********************************************************************************

  Returns: ``sdk.Result``

  Handles the logic behind the declaration of a new candidate.

``handleMsgEditCandidacy(ctx sdk.Context, msg MsgEditCandidacy, k Keeper)``
***************************************************************************

  Returns: ``sdk.Result``

  Handles the logic of the edition of an existing candidate.

``handleMsgDelegate(ctx sdk.Context, msg MsgDelegate, k Keeper)``
*****************************************************************

  Returns: ``sdk.Result``

  Handles the logic behind the delegation of shares to a validator candidate.

``handleMsgUnbond(ctx sdk.Context, msg MsgUnbond, k Keeper)``
*************************************************************

  Returns: ``sdk.Result``

  Handles the logic behind the unbonding of a delegator's bond from a validator candidate.
