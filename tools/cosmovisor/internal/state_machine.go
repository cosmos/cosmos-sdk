package internal

type Event interface{}

type UpgradePlanEvent struct{}
type UpgradeBatchEvent struct{}
type ProcessExistEvent struct{}
type Height struct{}

type StateMachine struct {
}

type State interface {
	OnEvent(event Event) State
}

//func (sm *StateMachine) WaitingForXUpgrade(event Event) {
//
//}
//
//func (sm *StateMachine) WaitingForManualUpgrade(event Event) {
//
//}

//	type WaitingForManualUpgrade struct {
//		manualUpgradePlan *cosmovisor.ManualUpgradePlan
//	}
//
// type WaitingForXUpgrade struct{}
//
// type WaitingForShutdown struct{}
func Init() {
	// check if have upgrade-info.json
	// check if have upgrade-info.json.batch
}

func WaitingForXUpgrade(event Event) {
	switch _ := event.(type) {
	case UpgradePlanEvent:
		// handle received upgrade plan event
	case UpgradeBatchEvent:
		// handle received upgrade batch event
	case ProcessExistEvent:
		// check for upgrade-info.json
	}
}
