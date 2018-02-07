package version

const Maj = "0"
const Min = "7"
const Fix = "0"

var (
	Version = "0.7.0"

	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
