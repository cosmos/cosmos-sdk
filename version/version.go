//nolint
package version

// when updating these,
// remember to also update examples/basecoin/tests/cli/rpc.sh
// TODO improve

const Maj = "0"
const Min = "7"
const Fix = "1"

var (
	Version = "0.7.1"

	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
