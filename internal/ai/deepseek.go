package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
)

type DeepSeekProvider struct {
	client *http.Client
	apiKey string
	model  string
}

func NewDeepSeekProvider(client *http.Client, apiKey, model string) *DeepSeekProvider {
	return &DeepSeekProvider{client: client, apiKey: strings.TrimSpace(apiKey), model: strings.TrimSpace(model)}
}

func (p *DeepSeekProvider) Name() string {
	return "deepseek"
}

func (p *DeepSeekProvider) AnalyzeRequest(text string, now time.Time) (domain.AnalyzeFreeformResponse, error) {
	var response domain.AnalyzeFreeformResponse
	err := p.generateJSON(providerAnalyzePrompt(text, now), &response)
	return response, err
}

func (p *DeepSeekProvider) EvaluateRequest(request domain.PlanningRequest) (domain.EvaluateRequestResponse, error) {
	var response domain.EvaluateRequestResponse
	err := p.generateJSON(providerEvaluatePrompt(request), &response)
	return response, err
}

func (p *DeepSeekProvider) GeneratePlan(request domain.PlanningRequest) (domain.GeneratedPlan, error) {
	var response domain.GeneratedPlan
	err := p.generateJSON(providerPlanPrompt(request), &response)
	return response, err
}

func (p *DeepSeekProvider) generateJSON(prompt string, target any) error {
	if p.apiKey == "" {
		return fmt.Errorf("DEEPSEEK_API_KEY is required")
	}
	if p.client == nil {
		p.client = &http.Client{Timeout: 45 * time.Second}
	}
	model := p.model
	if model == "" {
		model = "deepseek-chat"
	}

	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "Return only valid JSON that matches the requested schema. Do not include markdown."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return fmt.Errorf("deepseek API error: %s", decoded.Error.Message)
		}
		return fmt.Errorf("deepseek API error: status %d", resp.StatusCode)
	}
	if len(decoded.Choices) == 0 || decoded.Choices[0].Message.Content == "" {
		return fmt.Errorf("deepseek API returned no content")
	}

	jsonPayload, err := extractJSONObject(decoded.Choices[0].Message.Content)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonPayload, target)
}
