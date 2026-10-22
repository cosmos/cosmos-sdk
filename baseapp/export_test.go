package baseapp

// MarkInternalForTest marks a Msg type URL as internal-only. Test-only helper.
func (msr *MsgServiceRouter) MarkInternalForTest(typeURL string) {
	msr.internalMsgs[typeURL] = struct{}{}
}
