Feature: Create group
  Scenario: Create a group from a single key
    Given a public key address
    When a user creates a group with that address and a decision threshold of 1
    Then they should get a new group address back
    And be able to retrieve the group details with that address
    And authorization should succeed with only there vote
    And authorization should fail with no votes
    And authorization should fail with any other votes
