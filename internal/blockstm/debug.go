package blockstm

import (
	"fmt"
	"time"
)

type sequenceDebugging struct {
	sequencing []*sequenceData
}

type sequenceData struct {
	start       time.Time
	end         time.Time
	suspensions []*suspendData
}

type suspendData struct {
	suspend time.Time
	resume  time.Time
}

func (s *Scheduler) DumpSequencing() {
	for index, data := range s.sequencing {
		if data == nil {
			fmt.Printf("txn %d: nil\n", index)
		}
		fmt.Printf("txn %d:\nstart: %v\nend: %v\n", index, data.start, data.end)
		for _, suspension := range data.suspensions {
			if suspension == nil {
				continue
			}
			fmt.Printf("\tsuspended: %v\n\tresumed: %v\n", suspension.suspend, suspension.resume)
		}
	}
}
