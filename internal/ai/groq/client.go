package groq

import (
	"net/http"
	"time"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		baseURL: "https://api.groq.com/openai/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: model,
	}
}

// Structs for http request messages and responses
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

type ResponseBody struct {
	Choices []Choice `json:"choices"`
}

func (c *GroqClient) AnalyzeTestFailure(prompt string) (string, error) {
	// Builds the GROQ url
	groqURL := c.baseURL + "/chat/completions"

	// Formats request body to match GROQ's API
	requestBody := RequestBody{
		Model: c.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	// Converts the request body into a JSON byte array
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Builds the post request using GROQ's API and sets the headers
	req, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + c.apiKey)

	// Sends the request and reads response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Reads the response body into a byte array
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parses and logs the response
	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil || len(response.Choices) == 0 {
		return "", fmt.Errorf("Failed to parse response or no result: %s", string(body))
	}
	return response.Choices[0].Message.Content, nil
}
