package planinfo

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PlanInfoTestSuite struct {
	suite.Suite
}

func TestPlanInfoTestSuite(t *testing.T) {
	suite.Run(t, new(PlanInfoTestSuite))
}

func (s PlanInfoTestSuite) TestParsePlanInfo() {
	// TODO: Unit tests for ParsePlanInfo(infoStr)
	// Tests:
	// * The infoStr is JSON:
	//   * Positive test: `{"binaries":{"os1/arch1":"url1","os2/arch2":"url2"}}`
	//   * Wrong data type: `{"binaries":["foo"]}`
	//   * Wrong data type in binaries: `{"binaries":{"os1/arch1":1,"os2/arch2":2}}`
	// * The infoStr is a URL:
	//   * Not really sure how to test this.
	//     Maybe use a non-existent example and just check that an http get error is returned.
	//   * url without checksum works
	//   * url with checksum works
	//   * url with incorrect checksum doesn't work.
}

func (s PlanInfoTestSuite) TestPlanInfoValidateFull() {
	// TODO: Unit tests for PlanInfo.ValidateFull(daemonName)
	// Tests:
	// * Check that a bad os/arch returns an error.
	// * Check that a download failure returns an error.
	// * Don't worry about a positive case.
}

func (s PlanInfoTestSuite) TestBinaryDownloadURLMapValidateBasic() {
	// TODO: Unit tests for BinaryDownloadURLMap.ValidateBasic()
	// Tests:
	// * An empty map returns an error.
	// * An empty string key returns an error.
	// * The "any" key does not return an error.
	// * A key of "xxx/yyy" does not return an error.
	// * A non-url value returns an error.
	// * A url without a checksum is valid.
	// * A url with a checksum is valid.
}

func (s PlanInfoTestSuite) TestBinaryDownloadURLMapCheckURLs() {
	// TODO: Unit tests for TestBinaryDownloadURLMap.CheckURLs(daemonName)
	// Not really sure how exactly to test this due to the http requests.
	// Tests:
	// * Non-existent URL returns an error.
	// * Maybe skip a positive test here.
	// * A url without a checksum works
	// * A url with a checksum works
	// * A url with an incorrect checksum returns an error.
}
