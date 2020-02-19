<!--
order: 13
synopsis: This document outlines the recommended usage and APIs for error handling in Cosmos SDK modules.
-->

# Errors

Modules are encouraged to define and register their own errors to provide better
context on failed message or handler execution. Typically, these errors should be
common or general errors which can be further wrapped to provide additional specific
execution context.
