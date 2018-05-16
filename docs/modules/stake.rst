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

Messages

::

  type MsgDeclareCandidacy struct {
  	Description
  	CandidateAddr sdk.Address   `json:"address"`
  	PubKey        crypto.PubKey `json:"pubkey"`
  	Bond          sdk.Coin      `json:"bond"`
  }

::

  type MsgEditCandidacy struct {
    Description
    CandidateAddr sdk.Address `json:"address"`
  }

::

  type MsgDelegate struct {
    DelegatorAddr sdk.Address `json:"address"`
    CandidateAddr sdk.Address `json:"address"`
    Bond          sdk.Coin    `json:"bond"`
  }

::

  type MsgUnbond struct {
    DelegatorAddr sdk.Address `json:"address"`
    CandidateAddr sdk.Address `json:"address"`
    Shares        string      `json:"shares"`
  }

``NewMsgDeclareCandidacy(candidateAddr sdk.Address, pubkey crypto.PubKey, bond sdk.Coin, description Description)``

  Returns a new ``MsgDeclareCandidacy`` message.

``NewMsgEditCandidacy(candidateAddr sdk.Address, description Description)``

  Returns a new ``MsgEditCandidacy`` message.

``NewMsgDelegate(delegatorAddr, candidateAddr sdk.Address, bond sdk.Coin)``

  Returns a new ``MsgDelegate`` instance.

``NewMsgUnbond(delegatorAddr, candidateAddr sdk.Address, shares string)``

  Returns a new ``MsgUnbond`` struct.
