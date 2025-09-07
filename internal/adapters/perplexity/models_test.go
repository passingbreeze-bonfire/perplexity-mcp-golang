package perplexity

import (
	"testing"
	"time"
)

func TestAPIChatResponse_GetCreatedTime(t *testing.T) {
	tests := []struct {
		name     string
		created  int64
		expected time.Time
	}{
		{
			name:     "valid timestamp",
			created:  1703097600, // 2023-12-20 16:00:00 UTC
			expected: time.Unix(1703097600, 0),
		},
		{
			name:     "zero timestamp",
			created:  0,
			expected: time.Unix(0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := APIChatResponse{Created: tt.created}
			got := resp.GetCreatedTime()
			if !got.Equal(tt.expected) {
				t.Errorf("GetCreatedTime() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIChatResponse_GetContent(t *testing.T) {
	tests := []struct {
		name     string
		choices  []APIChoice
		expected string
	}{
		{
			name: "single choice with content",
			choices: []APIChoice{
				{
					Index: 0,
					Message: APIMessage{
						Role:    "assistant",
						Content: "Hello, world!",
					},
					FinishReason: "stop",
				},
			},
			expected: "Hello, world!",
		},
		{
			name: "multiple choices returns first",
			choices: []APIChoice{
				{
					Index: 0,
					Message: APIMessage{
						Role:    "assistant",
						Content: "First response",
					},
					FinishReason: "stop",
				},
				{
					Index: 1,
					Message: APIMessage{
						Role:    "assistant",
						Content: "Second response",
					},
					FinishReason: "stop",
				},
			},
			expected: "First response",
		},
		{
			name:     "no choices",
			choices:  []APIChoice{},
			expected: "",
		},
		{
			name: "empty content",
			choices: []APIChoice{
				{
					Index: 0,
					Message: APIMessage{
						Role:    "assistant",
						Content: "",
					},
					FinishReason: "stop",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := APIChatResponse{Choices: tt.choices}
			got := resp.GetContent()
			if got != tt.expected {
				t.Errorf("GetContent() = %q, want %q", got, tt.expected)
			}
		})
	}
}
