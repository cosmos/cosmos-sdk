Feature: MsgSubmitProposal

  A proposal can be submitted if:
  - the deposit is greater or equal than the MinDeposit param
  - all Msgs' type URLs are registered and have the gov account as unique signer ("cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

  Rule: the deposit must be greater or equal than the MinDeposit param

  Scenario: deposit lesser than MinDeposit
    Given a MinDeposit param set to "100stake" and MinInitialDepositRation set to "1"
    When alice submits a proposal with deposit "10stake"
    Then expect the error "minimum deposit is too small"

  Scenario: deposit greater than MinDeposit
    Given a MinDeposit param set to "100stake" and MinInitialDepositRation set to "1"
    When alice submits a proposal with deposit "1000stake"
    Then expect no error

  Rule: all Msgs are registered and have the gov account as unique signer

  Scenario: Msg with incorrect signer
    When alice submits a proposal with Msg
      """
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "foobar",
        "to_address": "cosmos1lffv859v0fwg2tedwygdfux55l6xhmnt8xspcu",
        "amount": [{"denom": "stake", "amount": "10"}]
      }
      """
    Then expect the error "invalid bech32 string"
    
  Scenario: Msg with signer not equal to gov account
    When alice submits a proposal with Msg
      """
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "cosmos1lffv859v0fwg2tedwygdfux55l6xhmnt8xspcu",
        "to_address": "cosmos1lffv859v0fwg2tedwygdfux55l6xhmnt8xspcu",
        "amount": [{"denom": "stake", "amount": "10"}]
      }
      """
    Then expect the error "expected gov account as only signer for proposal message"

  Scenario: correct Msg with signer equal to gov account
    When alice submits a proposal with Msg
      """
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
        "to_address": "cosmos1lffv859v0fwg2tedwygdfux55l6xhmnt8xspcu",
        "amount": [{"denom": "stake", "amount": "10"}]
      }
      """
    Then expect no error

  Scenario: Msg with unknonwn TypeURL
    When alice submits a proposal with Msg
      """
      {
        "@type": "/testdata.TestMsg",
        "signers": ["cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"]
      }
      """
    Then expect the error "/testdata.TestMsg: proposal message not recognized by router"
