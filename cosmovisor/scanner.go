package cosmovisor

import (
	"bufio"
	"regexp"
)

// Trim off whitespace around the info - match least greedy, grab as much space on both sides
// Defined here: https://github.com/cosmos/cosmos-sdk/blob/release/v0.38.2/x/upgrade/abci.go#L38
//  fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
// DueAt defined here: https://github.com/cosmos/cosmos-sdk/blob/release/v0.38.2/x/upgrade/internal/types/plan.go#L73-L78
//
//    if !p.Time.IsZero() {
//      return fmt.Sprintf("time: %s", p.Time.UTC().Format(time.RFC3339))
//    }
//    return fmt.Sprintf("height: %d", p.Height)
var upgradeRegex = regexp.MustCompile(`UPGRADE "(.*)" NEEDED at ((height): (\d+)|(time): (\S+)):\s+(\S*)`)

// UpgradeInfo is the details from the regexp
type UpgradeInfo struct {
	Name string
	Info string
}

// WaitForUpdate will listen to the scanner until a line matches upgradeRegexp.
// It returns (info, nil) on a matching line
// It returns (nil, err) if the input stream errored
// It returns (nil, nil) if the input closed without ever matching the regexp
func WaitForUpdate(scanner *bufio.Scanner) (*UpgradeInfo, error) {
	for scanner.Scan() {
		line := scanner.Text()
		if upgradeRegex.MatchString(line) {
			subs := upgradeRegex.FindStringSubmatch(line)
			info := UpgradeInfo{
				Name: subs[1],
				Info: subs[7],
			}
			return &info, nil
		}
	}
	return nil, scanner.Err()
}
