Feature: inserting, updating and saving entities

  Scenario: can't insert an entity with a duplicate primary key
    Given an existing entity
      """
      {"name": "foo", "not_unique": "bar"}
      """
    When I insert
      """
      {"name": "foo", "not_unique": "baz"}
      """
    Then expect a "already exists" error
    And expect grpc error code "ALREADY_EXISTS"

  Scenario: can't update entity that doesn't exist
    When I update
        """
        {"name":"foo"}
        """
    Then expect a "not found" error
    And expect grpc error code "NOT_FOUND"
#
  Scenario: can't violate unique constraint on insert
    Given an existing entity
      """
      {"name": "foo", "unique": "bar"}
      """
    When I insert
      """
      {"name": "baz", "unique": "bar"}
      """
    Then expect a "unique key violation" error
    And expect grpc error code "FAILED_PRECONDITION"

  Scenario: can't violate unique constraint on update
    Given an existing entity
      """
      {"name": "foo", "unique": "bar"}
      """
    And an existing entity
      """
      {"name": "baz", "unique": "bam"}
      """
    When I update
      """
      {"name": "baz", "unique": "bar"}
      """
    Then expect a "unique key violation" error
    And expect grpc error code "FAILED_PRECONDITION"
