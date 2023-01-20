Feature: MsgTripCircuitBreaker

	Circuit breaker can disable message execution:
	- when the caller trips the circuitbreaker for a message or messages
	- when the permissions are valid

	Rule: the caller must have SUPER_ADMIN permission to trip the circuit
		Example: caller is a super-admin
			Given "acct1" has permission "LEVEL_SUPER_ADMIN"
			When  attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect success
		Example: caller has no permissions
			Given "acct1" has no permissions
			When "acct1" attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect an "unauthorized" error

	Rule: the caller must have ALL_MSGS permission to trip the circuit
		Example: caller has all msg permissions
			Given "acct1" has permission "LEVEL_ALL_MSGS"
			When "acct1" attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect an "unauthorized" error
		Example: caller has LEVEL_SOME_MSGS permissions
			Given "acct1" has permission "LEVEL_SOME_MSGS"
			When "acct1" attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect an "unauthorized" error
		Example: caller has no permissions
			Given "acct1" has no permissions
			When "acct1" attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect an "unauthorized" error

	Rule: the caller must have LEVEL_SOME_MSGS permission to trip the circuit
		Example: caller does not have permission for the message
			Given "acct1" has permission "LEVEL_SOME_MSGS"
			When "acct1" attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect an "unauthorized" error
		Example: caller has LEVEL_SOME_MSGS permissions
			Given "acct1" has permission "LEVEL_SOME_MSGS"
			When "acct1" attempts to disable msg execution
				"""
				{
					"msg": "cosmos.bank.v1beta1.MultiSend"
				}
				"""
			Then expect an "unauthorized" error
