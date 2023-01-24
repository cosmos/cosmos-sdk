Feature: MsgResetCircuitBreaker
	- Circuit breaker can be reset:
	- when the permissions are valid

  Rule: the caller has permission LEVEL_SUPER_ADMIN

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

    Example: caller attempts to reset a disabled message
      Given "acct1" has permission "LEVEL_SOME_MSGS"
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: caller has does not have permission for MultiSend
      Given "acct1" has permission "LEVEL_SOME_MSGS"
      When "acct1" attempts to reset a disabled message
        """
        {
        	"msg": "cosmos.bank.v1beta1.MultiSend"
        }
        """
      Then expect an "unauthorized" error
