package groq

import (
	"net/http"
	"time"
)

type GroqClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

func NewGroqClient(apiKey, model string) *GroqClient {
	return &GroqClient{
		apiKey:  apiKey,
		baseURL: "https://api.groq.com/openai/v1/chat/completions",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: model,
	}
}

func (c *GroqClient) AnalyzeTestFailure(prompt string) (string, error) {
	return "", nil
}
