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
	triggerUpgradeError           = "TriggerUpgradeError"
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

func haveWrongHaltHeight(_ context.Context, args ...any) bool {
	return false
}

func StateMachine(runner Runner) *stateless.StateMachine {
	fsm := stateless.NewStateMachine(computeRunPlan)

	// configure triggers for the state machine
	fsm.SetTriggerParameters(triggerGotUpgradeInfoJSON, reflect.TypeOf(&upgradetypes.Plan{}))
	fsm.SetTriggerParameters(triggerReadManualUpgradeBatch, reflect.TypeOf(cosmovisor.ManualUpgradeBatch{}))
	fsm.SetTriggerParameters(triggerReadLastKnownHeight, reflect.TypeOf(uint64(0)))
	fsm.SetTriggerParameters(triggerGotActualHeight, reflect.TypeOf(uint64(0)))
	fsm.SetTriggerParameters(triggerProcessExit, reflect.TypeOf(error(nil)))

	// configure ComputeRunPlan state
	fsm.Configure(computeRunPlan).
		InitialTransition(readUpgradeInfoJSON)

	fsm.Configure(readUpgradeInfoJSON).
		SubstateOf(computeRunPlan).
		OnEntry(runner.CheckForUpgradeInfoJSON).
		Permit(triggerGotUpgradeInfoJSON, readManualUpgradeBatch, isNil).
		Permit(triggerGotUpgradeInfoJSON, doUpgrade, isNotNil)

	fsm.Configure(readManualUpgradeBatch).
		SubstateOf(computeRunPlan).
		OnEntry(runner.CheckForManualUpgradeBatch).
		Permit(triggerReadManualUpgradeBatch, run, isNil).
		Permit(triggerReadManualUpgradeBatch, checkLastKnownHeight, isNotNil)

	fsm.Configure(checkLastKnownHeight).
		SubstateOf(computeRunPlan).
		OnEntry(runner.CheckLastKnownHeight).
		Permit(triggerReadLastKnownHeight, runWithHaltHeight, beforeManualUpgradeHeight).
		Permit(triggerReadLastKnownHeight, doUpgrade, atManualUpgradeHeight).
		Permit(triggerReadLastKnownHeight, fatalError, pastManualUpgradeHeight)

	// configure Run state

	fsm.Configure(run).
		OnEntry(runner.Start).
		Permit(triggerGotUpgradeInfoJSON, shutdownAndRestart).
		Permit(triggerGotNewManualUpgrade, shutdownAndRestart)

	// configure RunWithHaltHeight state

	fsm.Configure(runWithHaltHeight).
		Permit(triggerGotActualHeight, shutdownAndRestart, haveWrongHaltHeight).
		OnEntry(runner.StartWithHaltHeight).
		OnEntry(runner.CheckActualHeight).
		Permit(triggerGotUpgradeInfoJSON, shutdownAndRestart).
		Permit(triggerReachedHaltHeight, shutdownAndRestart).
		Permit(triggerGotNewManualUpgrade, shutdownAndRestart)

	// configure ShutdownAndRestart state
	fsm.Configure(shutdownAndRestart).
		OnEntry(runner.Stop).
		Permit(triggerProcessExit, computeRunPlan)

	// configure DoUpgrade state
	fsm.Configure(doUpgrade).
		Permit(triggerUpgradeSuccess, computeRunPlan).
		Permit(triggerUpgradeError, fatalError)

	return fsm

	//OnEntry(func(ctx context.Context, args ...any) error {
	//	// TODO read upgrade-info.json
	//	// if upgrade-info.json exists, read it and return the state
	//	panic("ReadUpgradeInfoJson state entered")
	//})
}

type Watcher int

type Runner interface {
	CheckForUpgradeInfoJSON(ctx context.Context, args ...any) error
	CheckForManualUpgradeBatch(ctx context.Context, args ...any) error
	CheckLastKnownHeight(ctx context.Context, args ...any) error
	CheckActualHeight(ctx context.Context, args ...any) error
	Start(ctx context.Context, args ...any) error
	StartWithHaltHeight(ctx context.Context, args ...any) error
	Stop(ctx context.Context, args ...any) error
}

//type MockRunner struct{}
//
//func (m MockRunner) CheckForUpgradeInfoJSON(ctx context.Context) error {
//	return nil
//}
//
////func (m MockRunner) CheckForManualUpgradeBatch(ctx context.Context) {}
////
////func (m MockRunner) CheckLastKnownHeight(ctx context.Context) {}
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
