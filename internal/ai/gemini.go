package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
)

type GeminiProvider struct {
	client *http.Client
	apiKey string
	model  string
}

func NewGeminiProvider(client *http.Client, apiKey, model string) *GeminiProvider {
	return &GeminiProvider{client: client, apiKey: strings.TrimSpace(apiKey), model: strings.TrimSpace(model)}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) AnalyzeRequest(text string, now time.Time) (domain.AnalyzeFreeformResponse, error) {
	var response domain.AnalyzeFreeformResponse
	err := p.generateJSON(geminiAnalyzePrompt(text, now), &response)
	return response, err
}

func (p *GeminiProvider) EvaluateRequest(request domain.PlanningRequest) (domain.EvaluateRequestResponse, error) {
	var response domain.EvaluateRequestResponse
	err := p.generateJSON(providerEvaluatePrompt(request), &response)
	return response, err
}

func (p *GeminiProvider) GeneratePlan(request domain.PlanningRequest) (domain.GeneratedPlan, error) {
	var response domain.GeneratedPlan
	err := p.generateJSON(geminiPlanPrompt(request), &response)
	return response, err
}

func (p *GeminiProvider) generateJSON(prompt string, target any) error {
	if p.apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is required")
	}
	if p.client == nil {
		p.client = &http.Client{Timeout: 45 * time.Second}
	}
	model := p.model
	if model == "" {
		model = "gemini-1.5-flash"
	}

	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/" + url.PathEscape(model) + ":generateContent?key=" + url.QueryEscape(p.apiKey)
	body := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"temperature":      0.2,
		},
	}

	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var decoded struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return fmt.Errorf("gemini API error: %s", decoded.Error.Message)
		}
		return fmt.Errorf("gemini API error: status %d", resp.StatusCode)
	}
	if len(decoded.Candidates) == 0 || len(decoded.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("gemini API returned no content")
	}

	jsonPayload, err := extractJSONObject(decoded.Candidates[0].Content.Parts[0].Text)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonPayload, target)
}

func geminiAnalyzePrompt(text string, now time.Time) string {
	return providerAnalyzePrompt(text, now)
}

func geminiPlanPrompt(request domain.PlanningRequest) string {
	return providerPlanPrompt(request)
}
