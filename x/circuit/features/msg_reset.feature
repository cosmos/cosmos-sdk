Feature: MsgResetCircuitBreaker
	- Circuit breaker can be reset:
	- when the permissions are valid

  Rule: caller must have a permission to reset the circuit

    Example: caller attempts to reset a disabled message
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      When "acct1" attempts to enable a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: caller has no permissions
      Given "acct1" has no permissions
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect an "unauthorized" error

    Example: caller attempts to reset a disabled message
      Given "acct1" has permission "LEVEL_ALL_MSGS"
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: caller attempts to reset a message they have permission to trip
      Given "acct1" has permission to trip circuit breaker for "cosmos.bank.v1beta1.MsgSend"
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: caller attempts to reset a message they don't have permission to trip
      Given "acct1" has permission to trip circuit breaker for "cosmos.bank.v1beta1.MsgSend"
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MultiSend"
        }
        """
      Then expect success

    Example: caller attempts to reset a message that has been tripped
      Given "acct1" has permission "LEVEL_SUPER_ADMIN" & "cosmos.bank.v1beta1.MultiSend" has been enabled
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MultiSend"
        }
        """
      Then expect an "msg enabled" error
