package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
)

func main() {
	for i := 0; i < 50; i++ {
		func() {
			f, err := os.Create(fmt.Sprintf("corpus/%d.seed", i+1))
			if err != nil {
				return
			}
			defer f.Close()
			rnd := rand.Intn(500000)
			ti := time.Now().Add(time.Duration(-rnd) * time.Hour)
			b := types.FormatTimeBytes(ti)
			f.Write(b)
		}()
	}
}
