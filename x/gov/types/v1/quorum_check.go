package v1

import (
	time "time"
)

func NewQuorumCheckQueueEntry(quorumTimeoutTime time.Time, quorumCheckCount uint64) QuorumCheckQueueEntry {
	return QuorumCheckQueueEntry{
		QuorumTimeoutTime: &quorumTimeoutTime,
		QuorumCheckCount:  quorumCheckCount,
		QuorumChecksDone:  0,
	}
}
