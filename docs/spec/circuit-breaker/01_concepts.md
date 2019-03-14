# Concepts

The intention of the circuit breaker is to have a contingency plan for a
running network which maintains network liveness. This can be achieved through
selectively "pausing" functionality of specific modules on a running network.
The circuit breaker is intended to be enabled through either:
 - governance, 
 - the bonded validator group (for emergencies), 
 - special transaction (which proves how expected behaviour is broken). 

## Pause state

The basic pause state of any module simple disables all message routes to
that module.  Beyond that, it may be a appropriate for different modules to
process begin-block/end-block in an altered "safe" way. 

