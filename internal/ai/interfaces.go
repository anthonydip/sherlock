package ai

type AIClient interface {
	AnalyzeTestFailure(prompt string) (string, error)
}
