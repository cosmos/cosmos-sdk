Feature: MsgWeightedVote

  A user can vote if:
  - the proposal is in voting period
  - the vote options are correct

  Rule: the proposal is in voting period

  Scenario: vote on a non-existing proposal yields error
    When alice votes "yes=1" on proposal 42
    Then expect the error "42: unknown proposal"

  Scenario: vote on inactive proposal (not enough deposit) yields error
    Given a proposal with "0stake" deposit
    When alice votes "yes=1" on proposal 1
    Then expect the error "1: inactive proposal"
  
  Rule: the vote options are correct

  Scenario: unknown vote option yields error
    Given a proposal with "10000000stake" deposit
    When alice votes "foo=1" on proposal 1
    Then expect the error "'foo' is not a valid vote option"

  Scenario Outline: correct votes
    Given a proposal with "10000000stake" deposit
    When alice votes "<vote-option>" on proposal 1
    Then expect no error

    Examples:
    | vote-option                          |
    | yes=1                                |
    | yes=1.0000                           |
    | yes=0.5,no=0.5                       |
    | abstain=0.00001,no_with_veto=0.99999 |

  Scenario: can vote twice on proposal
    Given a proposal with "10000000stake" deposit
    And alice votes "yes=1" on proposal 1
    When alice votes "no=1" on proposal 1
    Then expect no error
