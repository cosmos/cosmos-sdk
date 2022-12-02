Feature: invokers

  Invokers are functions the will always get called, have strictly optional
  dependencies and no return outputs (other than error).

  Background:

  Rule: invokers get called even if their dependencies can't be resolved
    Example: no providers
      Given an invoker requesting an int and string pointer
      When the container is built
      Then the invoker will get the int parameter set to 0
      And the invoker will get the string pointer parameter set to nil

  Rule: invokers get called with dependencies if they are provided
    Example: int and string pointer providers
      Given an invoker requesting an int and string pointer
      And an int provider returning 5
      And a string pointer provider pointing to "foo"
      When the container is built
      Then the invoker will get the int parameter set to 5
      And the invoker will get the string pointer parameter set to "foo"

  Rule: invokers get module scoped dependencies
    Example: module-scoped int
      Given an invoker requesting an int and string pointer run in module "foo"
      And a module-scoped int provider which returns the length of the module name
      When the container is built
      Then the invoker will get the int parameter set to 3
      And the invoker will get the string pointer parameter set to nil
