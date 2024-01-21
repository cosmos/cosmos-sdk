package types

import (
	"cosmossdk.io/math"
)

const (
	nanosecondsPerSecond = 1000000000
	secondsPerMinute     = 60
	minutesPerHour       = 60
	hoursPerDay          = 24
	daysPerYear          = 365.2425

	secondsPerYear     = int64(secondsPerMinute * minutesPerHour * hoursPerDay * daysPerYear)                        // 31,556,952
	nanosecondsPerYear = int64(nanosecondsPerSecond * secondsPerMinute * minutesPerHour * hoursPerDay * daysPerYear) // 31,556,952,000,000,000
)

var (
	nanosecondsPerYearDec = math.LegacyNewDec(nanosecondsPerYear)
)

var (
	initialInflationRate = math.LegacyNewDecWithPrec(8, 2)  // 0.08
	disinflationRate     = math.LegacyNewDecWithPrec(1, 1)  // 0.10
	targetInflationRate  = math.LegacyNewDecWithPrec(15, 3) // 0.015
)
