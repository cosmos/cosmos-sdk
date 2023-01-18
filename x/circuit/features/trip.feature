Feature: MsgTripCircuitBreaker

	Circuit breaker can disable message execution:
	- when the caller trips the circuitbreaker for a message or messages
	- when the permissions are valid

	Rule: the caller must have permission to trip to trip
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

		Example: caller has specific msg permissions
			Given "acct1" has permission "LEVEL_SOME_MSGS"
			When "acct1" attempts to disable msg execution for a message it doesnt have permission
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect an "unauthorized" error
