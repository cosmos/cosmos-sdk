Feature: interface type resolution

  Background:
    Given an interface Duck
    And two implementations Mallard and Canvasback

  Rule: interface types resolve to a concrete type implicitly if there is only one matching implementation
    Example: only one implementation
      Given "Mallard" is provided
      When we try to resolve a "Duck" in global scope
      Then "Mallard" is resolved in global scope

    Example: two implementations
      Given "Mallard" is provided
      * "Canvasback" is provided
      When we try to resolve a "Duck" in global scope
      Then there is a "Multiple implementations found" error

  Rule: preferences must point to a real type
    Example: a preferred type is not provided
      Given "Mallard" is provided
      And there is a global preference for a "Marbled" "Duck"
      When we try to resolve a "Duck" in global scope
      Then there is a "No type for explicit binding" error

  Rule: preferences supersede implicit type resolution
    Example: global scope
      Given "Canvasback" is provided
      And there is a global preference for a "Mallard" "Duck"
      When we try to resolve a "Duck" in global scope
      Then there is a "No type for explicit binding" error

    Example: module scope
      Given "Canvasback" is provided
      And there is a preference for a "Mallard" "Duck" in module "A"
      When module "A" wants a "Duck"
      Then there is a "No type for explicit binding" error

  Rule: preferences in global scope apply to both global and module-scoped resolution (if there is no module-scoped preference)
    Example: global resolution
      Given "Mallard" is provided
      And "Canvasback" is provided
      And there is a global preference for a "Mallard" "Duck"
      When we try to resolve a "Duck" in global scope
      Then "Mallard" is resolved in global scope

    Example: module-scoped resolution
      Given "Mallard" is provided
      And "Canvasback" is provided
      And there is a global preference for a "Mallard" "Duck"
      When module "A" wants a "Duck"
      Then module "A" resolves a "Mallard"

  Rule: module-scoped preferences only apply to module-scoped resolution
    Example: a module-scoped binding doesn't work for global scope
      Given "Mallard" is provided
      * "Canvasback" is provided
      * there is a preference for a "Canvasback" "Duck" in module "A"
      When we try to resolve a "Duck" in global scope
      Then there is a "Multiple implementations found" error

    Example: a module-scoped binding works for that module
      Given "Mallard" is provided
      * "Canvasback" is provided
      * there is a preference for a "Canvasback" "Duck" in module "A"
      When module "A" wants a "Duck"
      Then module "A" resolves a "Canvasback"

    Example: a module-scoped binding doesn't work for another module
      Given "Mallard" is provided
      * "Canvasback" is provided
      * there is a preference for a "Canvasback" "Duck" in module "A"
      When module "B" wants a "Duck"
      Then there is a "Multiple implementations found" error

    # this case is called a "journey" scenario which tests a bunch of things together
    # most tests should be short and to the point like the ones above but one or two long ones
    # are good to test more things together &/or do integration tests
    Example: two module-scoped preferences and a global preference
      Given "Mallard" is provided
      * "Canvasback" is provided
      * "Marbled" is provided
      * there is a global preference for a "Marbled" "Duck"
      * there is a preference for a "Canvasback" "Duck" in module "A"
      * there is a preference for a "Mallard" "Duck" in module "B"
      When module "A" wants a "Duck"
      * module "B" wants a "Duck"
      * module "C" wants a "Duck"
      * we try to resolve a "Duck" in global scope
      Then there is no error
      * module "A" resolves a "Canvasback"
      * module "B" resolves a "Mallard"
      * module "C" resolves a "Marbled"
      * "Marbled" is resolved in global scope
