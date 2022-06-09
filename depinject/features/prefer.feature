Feature: interface type resolution

  Background:
    Given an interface Duck
    And two implementations Mallard and Canvasback

  Rule: interface types resolve to a concrete type implicitly if there is only one matching implementation
  provided in the container
    Example: only one implementation
      Given "Mallard" is provided
      When we try to resolve a "Duck" in global scope
      Then "Mallard" is resolved

    Example: two implementations
      Given "Mallard" and "Canvasback" are provided
      When we try to resolve a "Duck" in global scope
      Then there is a "multiple implicit interface bindings" error

  Rule: preferences must point to a real type
    Example:
      Given "Mallard" is provided
      And there is a global preference for a "Marbled" "Duck"
      When we try to resolve a "Duck" in global scope
      Then there is a "no type for explicit binding" error

  Rule: preferences supersede implicit type resolution
    Example: global scope
      Given "Canvasback" is provided
      And there is a global preference for a "Mallard" "Duck"
      When we try to resolve a "Duck" in global scope
      Then there is a "can't resolve" error

    Example: module scope
      Given "Canvasback" is provided
      And there is a preference for a "Mallard" "Duck" in module "A"
      When module "A" wants a "Duck"
      Then there is a "can't resolve" error

  Rule: preferences in global scope apply to both global and module-scoped resolution, if there is no module-scoped preference
    Example: global resolution
      Given "Mallard" and "Canvasback" are provided
      And there is a global preference for a "Mallard" "Duck"
      When we try to resolve a "Duck" in global scope
      Then "Mallard" is resolved

    Example: module-scoped resolution
      Given "Mallard" and "Canvasback" are provided
      And there is a global preference for a "Mallard" "Duck"
      When module "A" wants a "Duck"
      Then module "A" resolves a "Mallard"

  Rule: module-scoped preferences only apply to module-scoped resolution
    Example: a module-scoped binding doesn't work for global scope
      Given "Mallard" and "Canvasback" are provided
      And there is a preference for a "Canvasback" "Duck" in module "A"
      When we try to resolve a "Duck" in global scope
      Then there is a "multiple implicit interface bindings" error

    Example: a module-scoped binding works for that module
      Given "Mallard" and "Canvasback" are provided
      And there is a preference for a "Canvasback" "Duck" in module "A"
      When module "A" wants a "Duck"
      Then module "A" resolves a "Canvasback"

    Example: a module-scoped binding doesn't work for another module
      Given "Mallard" and "Canvasback" are provided
      And there is a preference for a "Canvasback" "Duck" in module "A"
      When module "B" wants a "Duck"
      Then there is a "multiple implicit interface bindings" error

    Example: two module-scoped preferences and a global preference
      Given "Mallard" and "Canvasback" are provided
      * there is a global preference for a "Mallard" "Duck"
      * there is a preference for a "Canvasback" "Duck" in module "A"
      * there is a preference for a "Mallard" "Duck" in module "B"
      When module "A" wants a "Duck"
      * module "B" wants a "Duck"
      * module "C" wants a "Duck"
      Then module "A" resolves a "Canvasback"
      * module "B" resolves a "Mallard"
      * module "C" resolves a "Mallard"
