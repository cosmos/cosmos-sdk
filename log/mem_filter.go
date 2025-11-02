package log

import "strings"

// buildDefaultAllowedMsgs returns a lowercased allow-list for messages
// that should be recorded by memlogger when filtering is enabled. Matching is case-insensitive and uses
// exact string equality after lowercasing.
func buildDefaultAllowedMsgs() map[string]struct{} {
	msgs := []string{
		// Node identity
		"This node is a validator",
		"P2P Node ID",
		"Signed proposal",

		// Votes / p2p
		"Adding vote",
		"Added vote to prevote",
		"Added vote to precommit",
		"Added vote to last precommits",
		"Sending vote message",
		"Send",
		"Receive",
		"Read PacketMsg",
		"TrySend",
		"Received bytes",
		"Received proposal",
		"Receive block part",
		"Received complete proposal block",

		// Consensus
		"Entering new round",
		"Entering new round with invalid args",
		"Entering propose step",
		"Entering propose step with invalid args",
		"Propose step; our turn to propose",
		"Propose step; not our turn to propose",
		"Entering prevote step",
		"Entering prevote step with invalid args",
		"Entering prevote wait step",
		"Entering prevote wait step with invalid args",
		"Entering precommit step",
		"Entering precommit step with invalid args",
		"Entering precommit wait step",
		"Entering precommit wait step with invalid args",
		"Entering commit step",
		"Entering commit step with invalid args",
		"Finalizing commit of block",
		"Committed block",
		"Updating valid block because of POL",
		"Precommit step: +2/3 prevoted proposal block; locking",
		"Scheduled timeout",

		// Storage
		"Store working hash",
	}
	out := make(map[string]struct{}, len(msgs))
	for _, m := range msgs {
		out[strings.ToLower(m)] = struct{}{}
	}
	return out
}
