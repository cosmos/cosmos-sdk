package planinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PlanInfoTestSuite struct {
	suite.Suite

	// Home is a temporary directory for use in these tests.
	Home string
}

func (s *PlanInfoTestSuite) SetupTest() {
	s.Home = s.T().TempDir()
	s.T().Logf("Home: [%s]", s.Home)
}

func TestPlanInfoTestSuite(t *testing.T) {
	suite.Run(t, new(PlanInfoTestSuite))
}

// saveSrcTestFile saves a TestFile in this test's Home/src directory.
// The full path to the saved file is returned.
func (s PlanInfoTestSuite) saveTestFile(f *TestFile) string {
	fullName, err := f.SaveIn(s.Home)
	s.Require().NoError(err, "saving test file %s", f.Name)
	return fullName
}

func (s PlanInfoTestSuite) TestParsePlanInfo() {
	goodJSON := `{"binaries":{"os1/arch1":"url1","os2/arch2":"url2"}}`
	binariesWrongJSON := `{"binaries":["foo","bar"]}`
	binariesWrongValueJSON := `{"binaries":{"os1/arch1":1,"os2/arch2":2}}`
	goodJSONPath := s.saveTestFile(NewTestFile("good.json", goodJSON))
	binariesWrongJSONPath := s.saveTestFile(NewTestFile("binaries-wrong.json", binariesWrongJSON))
	binariesWrongValueJSONPath := s.saveTestFile(NewTestFile("binaries-wrong-value.json", binariesWrongValueJSON))
	goodJSONPlanInfo := &PlanInfo{
		Binaries: BinaryDownloadURLMap{
			"os1/arch1": "url1",
			"os2/arch2": "url2",
		},
	}
	makeInfoStrFuncString := func(val string) func(t *testing.T) string {
		return func(t *testing.T) string {
			return val
		}
	}
	makeInfoStrFuncURL := func(file string) func(t *testing.T) string {
		return func(t *testing.T) string {
			return makeFileURL(t, file)
		}
	}

	tests := []struct {
		name             string
		infoStrMaker     func(t *testing.T) string
		expectedPlanInfo *PlanInfo
		expectedInError  []string
	}{
		{
			name:             "json good",
			infoStrMaker:     makeInfoStrFuncString(goodJSON),
			expectedPlanInfo: goodJSONPlanInfo,
			expectedInError:  nil,
		},
		{
			name:             "blank string",
			infoStrMaker:     makeInfoStrFuncString("   "),
			expectedPlanInfo: nil,
			expectedInError:  []string{"plan info cannot be blank"},
		},
		{
			name:             "json binaries is wrong data type",
			infoStrMaker:     makeInfoStrFuncString(binariesWrongJSON),
			expectedPlanInfo: nil,
			expectedInError:  []string{"could not parse plan info", "cannot unmarshal array into Go struct field PlanInfo.binaries"},
		},
		{
			name:             "json wrong data type in binaries value",
			infoStrMaker:     makeInfoStrFuncString(binariesWrongValueJSON),
			expectedPlanInfo: nil,
			expectedInError:  []string{"could not parse plan info", "cannot unmarshal number into Go struct field PlanInfo.binaries"},
		},
		{
			name:             "url does not exist",
			infoStrMaker:     makeInfoStrFuncString("file:///this/file/does/not/exist?checksum=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
			expectedPlanInfo: nil,
			expectedInError:  []string{"could not download url", "file:///this/file/does/not/exist"},
		},
		{
			name:             "url good",
			infoStrMaker:     makeInfoStrFuncURL(goodJSONPath),
			expectedPlanInfo: goodJSONPlanInfo,
			expectedInError:  nil,
		},
		{
			name:             "url binaries is wrong data type",
			infoStrMaker:     makeInfoStrFuncURL(binariesWrongJSONPath),
			expectedPlanInfo: nil,
			expectedInError:  []string{"could not parse plan info", "cannot unmarshal array into Go struct field PlanInfo.binaries"},
		},
		{
			name:             "url wrong data type in binaries value",
			infoStrMaker:     makeInfoStrFuncURL(binariesWrongValueJSONPath),
			expectedPlanInfo: nil,
			expectedInError:  []string{"could not parse plan info", "cannot unmarshal number into Go struct field PlanInfo.binaries"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			infoStr := tc.infoStrMaker(t)
			actualPlanInfo, actualErr := ParsePlanInfo(infoStr)
			if len(tc.expectedInError) > 0 {
				require.Error(t, actualErr)
				for _, expectedErr := range tc.expectedInError {
					assert.Contains(t, actualErr.Error(), expectedErr)
				}
			} else {
				require.NoError(t, actualErr)
			}
			assert.Equal(t, tc.expectedPlanInfo, actualPlanInfo)
		})
	}
}

func (s PlanInfoTestSuite) TestPlanInfoValidateFull() {
	darwinAMD64File := NewTestFile("darwin_amd64", "#!/usr/bin\necho 'darwin/amd64'\n")
	linux386File := NewTestFile("linux_386", "#!/usr/bin\necho 'darwin/amd64'\n")
	darwinAMD64Path := s.saveTestFile(darwinAMD64File)
	linux386Path := s.saveTestFile(linux386File)
	darwinAMD64URL := makeFileURL(s.T(), darwinAMD64Path)
	linux386URL := makeFileURL(s.T(), linux386Path)

	tests := []struct {
		name     string
		planInfo *PlanInfo
		errs     []string
	}{
		// Positive test case
		{
			name: "two good entries",
			planInfo: &PlanInfo{
				Binaries: BinaryDownloadURLMap{
					"darwin/amd64": darwinAMD64URL,
					"linux/386":    linux386URL,
				},
			},
			errs: nil,
		},
		// a failure from BinaryDownloadURLMap.ValidateBasic
		{
			name:     "empty binaries",
			planInfo: &PlanInfo{Binaries: BinaryDownloadURLMap{}},
			errs:     []string{"no \"binaries\" entries found"},
		},
		// a failure from BinaryDownloadURLMap.CheckURLS
		{
			name: "url does not exist",
			planInfo: &PlanInfo{
				Binaries: BinaryDownloadURLMap{
					"darwin/arm64": "file:///no/such/file/exists/hopefully.zip?checksum=sha256:b5a2c96250612366ea272ffac6d9744aaf4b45aacd96aa7cfcb931ee3b558259",
				},
			},
			errs: []string{"error downloading binary", "darwin/arm64", "no such file or directory"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actualErr := tc.planInfo.ValidateFull("daemon")
			if len(tc.errs) > 0 {
				require.Error(t, actualErr)
				for _, expectedErr := range tc.errs {
					assert.Contains(t, actualErr.Error(), expectedErr)
				}
			} else {
				require.NoError(t, actualErr)
			}
		})
	}
}

func (s PlanInfoTestSuite) TestBinaryDownloadURLMapValidateBasic() {
	addDummyChecksum := func(url string) string {
		return url + "?checksum=sha256:b5a2c96250612366ea272ffac6d9744aaf4b45aacd96aa7cfcb931ee3b558259"
	}
	tests := []struct {
		name   string
		urlMap BinaryDownloadURLMap
		errs   []string
	}{
		{
			name:   "empty map",
			urlMap: BinaryDownloadURLMap{},
			errs:   []string{"no \"binaries\" entries found"},
		},
		{
			name: "key with empty string",
			urlMap: BinaryDownloadURLMap{
				"": addDummyChecksum("https://v1.cosmos.network/sdk"),
			},
			errs: []string{"invalid os/arch", `""`},
		},
		{
			name: "invalid key format",
			urlMap: BinaryDownloadURLMap{
				"badkey": addDummyChecksum("https://v1.cosmos.network/sdk"),
			},
			errs: []string{"invalid os/arch", "badkey"},
		},
		{
			name: "any key is valid",
			urlMap: BinaryDownloadURLMap{
				"any": addDummyChecksum("https://v1.cosmos.network/sdk"),
			},
			errs: nil,
		},
		{
			name: "os arch key is valid",
			urlMap: BinaryDownloadURLMap{
				"darwin/amd64": addDummyChecksum("https://v1.cosmos.network/sdk"),
			},
			errs: nil,
		},
		{
			name: "not a url",
			urlMap: BinaryDownloadURLMap{
				"isa/url":  addDummyChecksum("https://v1.cosmos.network/sdk"),
				"nota/url": addDummyChecksum("https://v1.cosmos.network:not-a-port/sdk"),
			},
			errs: []string{"invalid url", "nota/url", "invalid port"},
		},
		{
			name: "url without checksum",
			urlMap: BinaryDownloadURLMap{
				"darwin/amd64": "https://v1.cosmos.network/sdk",
			},
			errs: []string{"invalid url", "darwin/amd64", "missing checksum query parameter"},
		},
		{
			name: "multiple valid entries but one bad url",
			urlMap: BinaryDownloadURLMap{
				"any":          addDummyChecksum("https://v1.cosmos.network/sdk"),
				"darwin/amd64": addDummyChecksum("https://v1.cosmos.network/sdk"),
				"darwin/arm64": addDummyChecksum("https://v1.cosmos.network/sdk"),
				"windows/bad":  addDummyChecksum("https://v1.cosmos.network:not-a-port/sdk"),
				"linux/386":    addDummyChecksum("https://v1.cosmos.network/sdk"),
			},
			errs: []string{"invalid url", "windows/bad", "invalid port"},
		},
		{
			name: "multiple valid entries but one bad key",
			urlMap: BinaryDownloadURLMap{
				"any":          addDummyChecksum("https://v1.cosmos.network/sdk"),
				"darwin/amd64": addDummyChecksum("https://v1.cosmos.network/sdk"),
				"badkey":       addDummyChecksum("https://v1.cosmos.network/sdk"),
				"darwin/arm64": addDummyChecksum("https://v1.cosmos.network/sdk"),
				"linux/386":    addDummyChecksum("https://v1.cosmos.network/sdk"),
			},
			errs: []string{"invalid os/arch", "badkey"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actualErr := tc.urlMap.ValidateBasic()
			if len(tc.errs) > 0 {
				require.Error(t, actualErr)
				for _, expectedErr := range tc.errs {
					assert.Contains(t, actualErr.Error(), expectedErr)
				}
			} else {
				require.NoError(t, actualErr)
			}
		})
	}
}

func (s PlanInfoTestSuite) TestBinaryDownloadURLMapCheckURLs() {
	darwinAMD64File := NewTestFile("darwin_amd64", "#!/usr/bin\necho 'darwin/amd64'\n")
	linux386File := NewTestFile("linux_386", "#!/usr/bin\necho 'darwin/amd64'\n")
	darwinAMD64Path := s.saveTestFile(darwinAMD64File)
	linux386Path := s.saveTestFile(linux386File)
	darwinAMD64URL := makeFileURL(s.T(), darwinAMD64Path)
	linux386URL := makeFileURL(s.T(), linux386Path)

	tests := []struct {
		name   string
		urlMap BinaryDownloadURLMap
		errs   []string
	}{
		{
			name: "two good entries",
			urlMap: BinaryDownloadURLMap{
				"darwin/amd64": darwinAMD64URL,
				"linux/386":    linux386URL,
			},
			errs: nil,
		},
		{
			name: "url does not exist",
			urlMap: BinaryDownloadURLMap{
				"darwin/arm64": "file:///no/such/file/exists/hopefully.zip?checksum=sha256:b5a2c96250612366ea272ffac6d9744aaf4b45aacd96aa7cfcb931ee3b558259",
			},
			errs: []string{"error downloading binary", "darwin/arm64", "no such file or directory"},
		},
		{
			name: "bad checksum",
			urlMap: BinaryDownloadURLMap{
				"darwin/amd64": "file://" + darwinAMD64Path + "?checksum=sha256:b5a2c96250612366ea272ffac6d9744aaf4b45aacd96aa7cfcb931ee3b558259",
			},
			errs: []string{"error downloading binary", "darwin/amd64", "Checksums did not match", "b5a2c96250612366ea272ffac6d9744aaf4b45aacd96aa7cfcb931ee3b558259"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actualErr := tc.urlMap.CheckURLs("daemon")
			if len(tc.errs) > 0 {
				require.Error(t, actualErr)
				for _, expectedErr := range tc.errs {
					assert.Contains(t, actualErr.Error(), expectedErr)
				}
			} else {
				require.NoError(t, actualErr)
			}
		})
	}
}
