Feature: MsgSubmitProposal

  A proposal can be submitted:
  - the deposit is greater or equal than the MinDeposit param

  Rule: The deposit must be greater or equal than the MinDeposit param

  Scenario: deposit lesser than MinDeposit
    Given a MinDeposit param set to "100stake" and MinInitialDepositRation set to "1"
    When alice submits a proposal with deposit "10stake"
    Then expect the error "minimum deposit is too small"

  Scenario: deposit greater than MinDeposit
    Given a MinDeposit param set to "100stake" and MinInitialDepositRation set to "1"
    When alice submits a proposal with deposit "1000stake"
    Then expect no error
