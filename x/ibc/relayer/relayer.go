package relayer


// Relay implements the algorithm described in ICS18 (https://github.com/cosmos/ics/tree/master/spec/ics-018-relayer-algorithms)
func Relay(chains []Chain, strategy string) {
	for _, src := range chains {
		for _, dstID := range src.Counterparties {
			if dstID != src.Context.ChainID {
				dst := GetChain(dstID, chains)
				var msgs RelayMsgs

				// NOTE: This implemenation will allow for multiple strategies to be implemented
				// w/in this package and switched via config or flag
				if Strategy(strategy) != nil {
					msgs = Strategy(strategy)(src, dst)
				}

				// Submit the transactions to each chain
				src.SendMsgs(msgs.Src)
				dst.SendMsgs(msgs.Dst)
			}
		}
	}
}
