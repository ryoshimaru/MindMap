package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

func extractJSONObject(text string) ([]byte, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("empty AI response")
	}

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end < start {
		return nil, fmt.Errorf("AI response does not contain a JSON object")
	}

	payload := []byte(trimmed[start : end+1])
	if !json.Valid(payload) {
		return nil, fmt.Errorf("AI response JSON is invalid")
	}
	return payload, nil
}
