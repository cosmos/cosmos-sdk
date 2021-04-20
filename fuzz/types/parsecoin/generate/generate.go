package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	seeds := []struct {
		name string
		coin string
	}{
		{"valid", "10stake"},
		{"space_between_amount_denom", "10 stake"},
		{"leading_space", "  10stake"},
		{"trailing_space", "10stake  "},
		{"plus_sign", "+10stake"},
		{"minus_sign", "-10stake"},
		{"no_denom", "10"},
		{"less_than_3char_denom", "10s"},
		{"greater_than_128char_denom", "10" + strings.Repeat("s", 129)},
	}

	for _, seed := range seeds {
		func() {
			f, err := os.Create(fmt.Sprintf("corpus/%s.seed", seed.name))
			if err != nil {
				return
			}
			if _, err := f.Write([]byte(seed.coin)); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}
}
