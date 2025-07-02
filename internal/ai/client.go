package ai

import (
	"fmt"

	"github.com/anthonydip/sherlock/internal/ai/groq"
)

type AIOptions struct {
	Provider string
	Model    string
	APIKey   string
}

func NewAIClient(opts AIOptions) (AIClient, error) {
	switch opts.Provider {
	case "groq":
		return groq.NewGroqClient(opts.APIKey, opts.Model), nil
	default:
		return nil, fmt.Errorf("Unsupported AI client type: %s", opts.Provider)
	}
}
