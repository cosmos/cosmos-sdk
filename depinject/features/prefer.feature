Feature: interface type resolution

  Background:
    Given an interface Duck
    And two implementations Mallard and Canvasback

  Rule: interface types can be resolved to a concrete type implicitly if there is only one matching implementation
  provided in the container
    Scenario: only one implementation
      Given "Mallard" is provided
      When we try to resolve a "Duck" in global scope
      Then "Mallard" is resolved

    Scenario: two implementations
      Given "Mallard" and "Canvasback" are provided
      When we try to resolve a "Duck"
      Then there is an "multiple implicit interface bindings" error

  Scenario: two module-scoped preferences and a global preference
    Given "Mallard" and "Canvasback" are provided
    * there is a global preference for "Mallard"
    * there is a preference for "Canvasback" in module "A"
    * there is a preference for "Mallard" in module "B"
    When module "A" wants a "Duck"
    * module "B" wants a "Duck"
    * module "C" wants a "Duck"
    Then module "A" resolves a "Canvasback"
    * module "B" resolves a "Mallard"
    * module "C" resolves a "Mallard"
