package domain

import (
	"encoding/json"
	"testing"
)

func TestActivityTrackerUnmarshalConfidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence string
		want       int
	}{
		{name: "fraction", confidence: "0.9", want: 90},
		{name: "percentage", confidence: "90", want: 90},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tracker ActivityTracker
			if err := json.Unmarshal([]byte(`{"stage":"analysis_ready","confidence":`+test.confidence+`,"items":[]}`), &tracker); err != nil {
				t.Fatalf("unmarshal activity tracker: %v", err)
			}
			if tracker.Confidence != test.want {
				t.Fatalf("confidence = %d, want %d", tracker.Confidence, test.want)
			}
		})
	}
}
