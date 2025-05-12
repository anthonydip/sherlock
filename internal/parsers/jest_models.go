package parsers

type JestTestOutput struct {
	NumFailedTestSuites       int         `json:"numFailedTestSuites"`
	NumFailedTests            int         `json:"numFailedTests"`
	NumPassedTestSuites       int         `json:"numPassedTestSuites"`
	NumPassedTests            int         `json:"numPassedTests"`
	NumPendingTestSuites      int         `json:"numPendingTestSuites"`
	NumPendingTests           int         `json:"numPendingTests"`
	NumRuntimeErrorTestSuites int         `json:"numRuntimeErrorTestSuites"`
	NumTodoTests              int         `json:"numTodoTests"`
	NumTotalTestSuites        int         `json:"numTotalTestSuites"`
	NumTotalTests             int         `json:"numTotalTests"`
	OpenHandles               []string    `json:"openHandles"`
	StartTime                 int64       `json:"startTime"`
	Success                   bool        `json:"success"`
	TestResults               []TestSuite `json:"testResults"`
	WasInterrupted            bool        `json:"wasInterrupted"`
}

type TestSuite struct {
	AssertionResults []AssertionResult `json:"assertionResults"`
	EndTime          int64             `json:"endTime"`
	Message          string            `json:"message"`
	Name             string            `json:"name"`
	StartTime        int64             `json:"startTime"`
	Status           string            `json:"status"`
	Summary          string            `json:"summary"`
}

type AssertionResult struct {
	AncestorTitles    []string        `json:"ancestorTitles"`
	Duration          int             `json:"duration"`
	FailureDetails    []FailureDetail `json:"failureDetails"`
	FailureMessages   []string        `json:"failureMessages"`
	FullName          string          `json:"fullName"`
	Invocations       int             `json:"invocations"`
	Location          interface{}     `json:"location"`
	NumPassingAsserts int             `json:"numPassingAsserts"`
	RetryReasons      []string        `json:"retryReasons"`
	Status            string          `json:"status"`
	Title             string          `json:"title"`
}

type FailureDetail struct {
	MatcherResult *MatcherResult `json:"matcherResult,omitempty"`
}

type MatcherResult struct {
	Message string `json:"message"`
	Pass    bool   `json:"pass"`
}
