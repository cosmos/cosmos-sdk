Feature: MsgWeightedVote

  A user can vote if:
  - the proposal is in voting period
  - the vote options are correct

  Rule: proposal must be in the voting period to vote

    Scenario: cannot vote on a non-existing proposal
      When alice votes "yes=1" on proposal 42
      Then expect the error "42: unknown proposal"

    Scenario: cannot vote on inactive proposal (in DEPOSIT_PERIOD)
      Given a proposal with "0stake" initial deposit
      When alice votes "yes=1" on proposal 1
      Then expect the error "1: inactive proposal"

  Rule: vote options must be valid

    Scenario: cannote vote with unknown vote option
      Given a proposal with "10000000stake" initial deposit
      When alice votes "foo=1" on proposal 1
      Then expect the error "'foo' is not a valid vote option"

    Scenario Outline: correct votes
      Given a proposal with "10000000stake" initial deposit
      When alice votes "<vote-option>" on proposal 1
      Then expect no error

      Examples:
        | vote-option                          |
        | yes=1                                |
        | yes=1.0000                           |
        | yes=0.5,no=0.5                       |
        | abstain=0.00001,no_with_veto=0.99999 |

    Scenario: 2nd vote on same proposal overwrites the 1st one
      Given a proposal with "10000000stake" initial deposit
      And alice votes "yes=1" on proposal 1
      When alice votes "no=1" on proposal 1
      Then expect no error
      And alice's vote on proposal 1 is 'option:VOTE_OPTION_NO weight:"1.000000000000000000"'
