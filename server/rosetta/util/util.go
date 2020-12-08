package util

import "time"

// TryFunction tries to run `f` a number of times waiting, returning the last
// error encountered if all try actions fail
func TryFunction(times int, wait time.Duration, f func() error) (err error) {
	for i := 0; i < times; i++ {
		err = f()
		if err == nil {
			return
		}
		time.Sleep(wait)
	}
	return
}
