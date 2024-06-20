package indexertesting

import (
	"pgregory.net/rapid"
)

type StateMachineTestOptions struct {
	MaxBlocks int
	Init      func(*rapid.T)
	Check     func(*rapid.T)
	Cleanup   func(*rapid.T)
}

//func AppStateMachineTest(appOpts AppSimulatorOptions, testOpts StateMachineTestOptions) func(*rapid.T) {
//	maxBlocks := testOpts.MaxBlocks
//	if maxBlocks == 0 {
//		maxBlocks = 10
//	}
//
//	return func(t *rapid.T) {
//		if testOpts.Init != nil {
//			testOpts.Init(t)
//		}
//
//		if appOpts.AppSchema == nil {
//			appOpts.AppSchema = schemagen.AppSchema.Draw(t, "AppSchema")
//		}
//
//		appSim := NewAppSimulator(t, appOpts)
//		appSim.Initialize()
//
//		numBlocks := rapid.IntRange(1, maxBlocks).Draw(t, "numBlocks")
//
//		for i := 0; i < numBlocks; i++ {
//			appSim.actionNewBlock(t)
//			if testOpts.Check != nil {
//				testOpts.Check(t)
//			}
//		}
//
//		if testOpts.Cleanup != nil {
//			testOpts.Cleanup(t)
//		}
//	}
//}
