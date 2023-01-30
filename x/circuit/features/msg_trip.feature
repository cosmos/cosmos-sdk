Feature: MsgTripCircuitBreaker
	Circuit breaker can disable message execution:
	- when the caller trips the circuitbreaker for a message(s)
	- when the caller has the correct permissions

  Rule: a user must have permission to trip the circuit breaker for a message(s)

    Example: user is a super admin
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      When "acct1" attempts to disable msg execution
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: user has no permissions
      Given "acct1" has no permissions
      When "acct1" attempts to disable msg execution
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect an "unauthorized" error

    Example: user has permission for all messages
      Given "acct1" has permission "LEVEL_ALL_MSGS"
      When "acct1" attempts to disable msg execution
        """
        {
        	"msg": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: user has permission for the messages
      Given "acct1" has permission to disable "cosmos.bank.v1beta1.MsgSend" and "cosmos.staking.v1beta1.MsgDelegate"
      When "acct1" attempts to disable msg execution
        """
        {
        "msgs": ["cosmos.bank.v1beta1.MsgSend",cosmos.staking.v1beta1.MsgDelegate"]
        }
        """
      Then expect success

    Example: user does not have permission for 1 of the messages in the list
      Given "acct1" has permission to disable "cosmos.bank.v1beta1.MsgSend"
      When "acct1" attempts to disable msg execution
        """
        {
        "msgs": ["cosmos.bank.v1beta1.MsgSend","cosmos.staking.v1beta1.MsgCreateValidator"]
        }
        """
      Then expect an "unauthorized" error

    Example: user does not have permission for the message
      Given "acct1" has permission to diable "cosmos.bank.v1beta1.MsgSend"
      When "acct1" attempts to disable msg execution
        """
        {
        	"msg": "cosmos.bank.v1beta1.MultiSend"
        }
        """
      Then expect an "unauthorized" error

    Example: user tries to trip an already tripped circuit breaker
      Given "acct1" has permission to diable "cosmos.bank.v1beta1.MsgSend" & is already tripped
      When "acct1" attempts to disable msg execution
        """
        {
        	"msg": "cosmos.bank.v1beta1.MultiSend"
        }
        """
      Then expect an "msg disabled" error
