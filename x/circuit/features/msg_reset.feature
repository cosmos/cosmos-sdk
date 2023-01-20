Feature: MsgResetCircuitBreaker

	- Circuit breaker can be reset:
	- when the permissions are valid

	Rule: the caller must have permission
		Example: caller is a super-admin
			Given "acct1" has permission "LEVEL_SUPER_ADMIN"
			When  attempts to enable a disabled message
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


		Example: caller has all msg's permissions
			Given "acct1" has permission "LEVEL_ALL_MSGS"
			When "acct1" attempts to reset a disabled message
				"""
				{
					"msg": "cosmos.bank.v1beta1.MsgSend"
				}
				"""
			Then expect success
