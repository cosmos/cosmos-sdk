Feature: MsgDeposit

  A user can deposit if:
  - the proposal is in deposit period

  Rule: proposal must be in the deposit or voting period

  Scenario: cannot deposit on a non-existing proposal
    When alice deposits "100stake" on proposal 42
    Then expect the error "42: unknown proposal"

  Scenario: can deposit on proposal in DEPOSIT_PERIOD
    Given a MinDeposit param set to "100stake" and MinInitialDepositRatio set to "0"
    And a proposal with "50stake" initial deposit
    When alice deposits "10stake" on proposal 1
    Then expect no error
  
  Scenario: can deposit on proposal in VOTING_PERIOD
    Given a MinDeposit param set to "100stake" and MinInitialDepositRatio set to "1"
    And a proposal with "1000stake" initial deposit
    When alice deposits "50stake" on proposal 1
    Then expect no error
  
  Scenario: can deposit multiple times on same proposal
    Given a MinDeposit param set to "100stake" and MinInitialDepositRatio set to "1"
    And a proposal with "100stake" initial deposit
    And alice deposits "50stake" on proposal 1
    When alice deposits "50stake" on proposal 1
    Then expect no error

  # TODO Scenario: cannot deposit on proposal past its deposit end time

  # TODO Scenario: cannot deposit on proposal past its voting period

  Rule: proposal must have enough deposit to be active

  Scenario: proposal without enough initial deposit is in DEPOSIT_PERIOD
    Given a MinDeposit param set to "100stake" and MinInitialDepositRatio set to "0"
    And a proposal with "50stake" initial deposit
    When we query proposal 1
    Then expect the proposal to have status "PROPOSAL_STATUS_DEPOSIT_PERIOD"

  Scenario: proposal without enough total deposit is in DEPOSIT_PERIOD
    Given a MinDeposit param set to "100stake" and MinInitialDepositRatio set to "0"
    And a proposal with "50stake" initial deposit
    And alice deposits "20stake" on proposal 1
    When we query proposal 1
    Then expect the proposal to have status "PROPOSAL_STATUS_DEPOSIT_PERIOD"
    And expect the proposal to have total deposit "70stake"

  Scenario: proposal with enough total deposit is in VOTING_PERIOD
    Given a MinDeposit param set to "100stake" and MinInitialDepositRatio set to "0"
    And a proposal with "50stake" initial deposit
    And alice deposits "70stake" on proposal 1
    When we query proposal 1
    Then expect the proposal to have status "PROPOSAL_STATUS_VOTING_PERIOD"
    And expect the proposal to have total deposit "120stake"

  Rule: all deposits must be transferred to the gov module account

  Scenario: initial deposit is transferred to the gov module account
    Given a proposal with "50stake" initial deposit
    Then expect the gov account to have "50stake"

  Scenario: further deposits are transferred to the gov module account
    Given a proposal with "50stake" initial deposit
    When alice deposits "70stake" on proposal 1
    Then expect the gov account to have "120stake"
