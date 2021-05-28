package cosmovisor

import (
	"bufio"
	"regexp"
)

// Trim off whitespace around the info - match least greedy, grab as much space on both sides
// Defined here: https://github.com/cosmos/cosmos-sdk/blob/cb66c99eab17d0763ea900d8d7bf2d970e4add22/x/upgrade/abci.go#L73-L75
//    return fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
// DueAt defined here: https://github.com/cosmos/cosmos-sdk/blob/cb66c99eab17d0763ea900d8d7bf2d970e4add22/x/upgrade/types/plan.go#L39-L41
//    return fmt.Sprintf("Height: %d", p.Height)
var upgradeRegex = regexp.MustCompile(`UPGRADE "(.*)" NEEDED at ((Height): (\d+)):\s+(\S*)`)

// UpgradeInfo is the details from the regexp
type UpgradeInfo struct {
	Name   string
	Height string
	Info   string
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
				Name:   subs[1],
				Height: subs[4],
				Info:   subs[5],
			}
			return &info, nil
		}
	}
	return nil, scanner.Err()
}
