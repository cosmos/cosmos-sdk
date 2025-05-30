package internal

import (
	"context"
	"reflect"

	"github.com/qmuntal/stateless"

	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// triggers
var (
	triggerGotUpgradeInfoJSON     = "TriggerGotUpgradeInfoJSON"
	triggerReadManualUpgradeBatch = "TriggerReadManualUpgradeBatch"
	triggerReadLastKnownHeight    = "TriggerReadLastKnownHeight"
	triggerGotActualHeight        = "TriggerGotActualHeight"
	triggerGotNewManualUpgrade    = "TriggerGotNewManualUpgrade"
	triggerReachedHaltHeight      = "TriggerReachedHaltHeight"
	triggerProcessExit            = "TriggerProcessExit"
	triggerUpgradeSuccess         = "TriggerUpgradeSuccess"
)

// states
var (
	computeRunPlan         = "ComputeRunPlan"
	readUpgradeInfoJSON    = "ReadUpgradeInfoJson"
	readManualUpgradeBatch = "ReadManualUpgradeBatch"
	checkLastKnownHeight   = "CheckHeight"
	doUpgrade              = "DoUpgrade"
	run                    = "Run"
	runWithHaltHeight      = "RunWithHaltHeight"
	confirmHaltHeight      = "ConfirmHaltHeight"
	watchForHaltHeight     = "WatchForHaltHeight"
	shutdownAndRestart     = "ShutdownAndRestart"
	fatalError             = "FatalError"
)

func isNil(_ context.Context, args ...any) bool {
	// read manual upgrade batch uf upgrade-info.json is nil
	return args[0] == nil
}

func isNotNil(_ context.Context, args ...any) bool {
	// read manual upgrade batch uf upgrade-info.json is nil
	return args[0] != nil
}

func beforeManualUpgradeHeight(_ context.Context, args ...any) bool {
	return false
}

func atManualUpgradeHeight(_ context.Context, args ...any) bool {
	return false
}

func pastManualUpgradeHeight(_ context.Context, args ...any) bool {
	return false
}

func haveCorrectHaltHeight(_ context.Context, args ...any) bool {
	return false
}

func haveWrongHaltHeight(_ context.Context, args ...any) bool {
	return false
}

func StateMachine(runner Runner) *stateless.StateMachine {
	fsm := stateless.NewStateMachine(computeRunPlan)

	// configure triggers for the state machine
	fsm.SetTriggerParameters(triggerGotUpgradeInfoJSON, reflect.TypeOf(&upgradetypes.Plan{}))
	fsm.SetTriggerParameters(triggerReadManualUpgradeBatch, reflect.TypeOf(cosmovisor.ManualUpgradeBatch{}))
	fsm.SetTriggerParameters(triggerReadLastKnownHeight, reflect.TypeOf(uint64(0)))
	fsm.SetTriggerParameters(triggerProcessExit, reflect.TypeOf(error(nil)))

	// configure ComputeRunPlan state
	fsm.Configure(computeRunPlan).
		InitialTransition(readUpgradeInfoJSON)

	fsm.Configure(readUpgradeInfoJSON).
		SubstateOf(computeRunPlan).
		Permit(triggerGotUpgradeInfoJSON, readManualUpgradeBatch, isNil).
		Permit(triggerGotUpgradeInfoJSON, doUpgrade, isNotNil).
		OnActive(runner.ReadUpgradeInfoJsonSync)

	fsm.Configure(readManualUpgradeBatch).
		SubstateOf(computeRunPlan).
		Permit(triggerReadManualUpgradeBatch, run, isNil).
		Permit(triggerReadManualUpgradeBatch, checkLastKnownHeight, isNotNil)

	fsm.Configure(checkLastKnownHeight).
		SubstateOf(computeRunPlan).
		Permit(triggerReadLastKnownHeight, runWithHaltHeight, beforeManualUpgradeHeight).
		Permit(triggerReadLastKnownHeight, doUpgrade, atManualUpgradeHeight).
		Permit(triggerReadLastKnownHeight, fatalError, pastManualUpgradeHeight)

	// configure Run state

	fsm.Configure(run).
		OnEntry(func(ctx context.Context, args ...any) error {
			err := runner.StartWatchers(ctx, UpgradeInfoJsonWatcher, ManualUpgradeBatchWatcher)
			if err != nil {
				return err
			}
			return runner.StartProcess(ctx)
		}).
		OnExit(func(ctx context.Context, args ...any) error {
			return runner.StopWatchers(ctx)
		}).
		Permit(triggerGotUpgradeInfoJSON, shutdownAndRestart).
		Permit(triggerGotNewManualUpgrade, shutdownAndRestart)

	// configure RunWithHaltHeight state

	fsm.Configure(runWithHaltHeight).
		InitialTransition(confirmHaltHeight)

	fsm.Configure(confirmHaltHeight).
		SubstateOf(runWithHaltHeight).
		Permit(triggerGotActualHeight, watchForHaltHeight, haveCorrectHaltHeight).
		Permit(triggerGotActualHeight, shutdownAndRestart, haveWrongHaltHeight)

	fsm.Configure(watchForHaltHeight).
		SubstateOf(runWithHaltHeight).
		Permit(triggerGotUpgradeInfoJSON, shutdownAndRestart).
		Permit(triggerReachedHaltHeight, shutdownAndRestart).
		Permit(triggerGotNewManualUpgrade, shutdownAndRestart)

	// configure ShutdownAndRestart state
	fsm.Configure(shutdownAndRestart).
		OnActive(runner.StopProcess).
		Permit(triggerProcessExit, computeRunPlan)

	// configure DoUpgrade state
	fsm.Configure(doUpgrade).
		Permit(triggerUpgradeSuccess, computeRunPlan)

	return fsm

	//OnEntry(func(ctx context.Context, args ...any) error {
	//	// TODO read upgrade-info.json
	//	// if upgrade-info.json exists, read it and return the state
	//	panic("ReadUpgradeInfoJson state entered")
	//})
}

type Watcher int

const (
	UpgradeInfoJsonWatcher Watcher = iota
	ManualUpgradeBatchWatcher
	HeightWatcher
)

type Runner interface {
	ReadUpgradeInfoJsonSync(ctx context.Context) error
	ReadManualUpgradeBatchSync(ctx context.Context)
	CheckHeightSync(ctx context.Context)
	StartWatchers(ctx context.Context, watchers ...Watcher) error
	StopWatchers(ctx context.Context) error
	StartProcess(ctx context.Context) error
	StopProcess(ctx context.Context) error
}

//type MockRunner struct{}
//
//func (m MockRunner) ReadUpgradeInfoJsonSync(ctx context.Context) error {
//	return nil
//}
//
////func (m MockRunner) ReadManualUpgradeBatchSync(ctx context.Context) {}
////
////func (m MockRunner) CheckHeightSync(ctx context.Context) {}
////
////func (m MockRunner) WatchUpgradeInfoJson(ctx context.Context) {}
////
////func (m MockRunner) WatchManualUpgradeBatch(ctx context.Context) {}
////
////var _ Runner = &MockRunner{}
//
////type Event interface{}
////
////type UpgradePlanEvent struct{}
////type UpgradeBatchEvent struct{}
////type ProcessExistEvent struct{}
////type Height struct{}
////
////type StateMachine struct {
////}
////
////type State interface {
////	OnEvent(event Event) State
////}
////
//////func (sm *StateMachine) WaitingForXUpgrade(event Event) {
//////
//////}
//////
//////func (sm *StateMachine) WaitingForManualUpgrade(event Event) {
//////
//////}
////
//////	type WaitingForManualUpgrade struct {
//////		manualUpgradePlan *cosmovisor.ManualUpgradePlan
//////	}
//////
////// type WaitingForXUpgrade struct{}
//////
////// type WaitingForShutdown struct{}
////func Init() {
////	// check if have upgrade-info.json
////	// check if have upgrade-info.json.batch
////}
////
////func WaitingForXUpgrade(event Event) {
////	switch _ := event.(type) {
////	case UpgradePlanEvent:
////		// handle received upgrade plan event
////	case UpgradeBatchEvent:
////		// handle received upgrade batch event
////	case ProcessExistEvent:
////		// check for upgrade-info.json
////	}
////}
