package domain

import (
	"errors"
	"testing"
)

func TestSearchRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request SearchRequest
		wantErr bool
		errType error
	}{
		{
			name: "valid request with default model",
			request: SearchRequest{
				Query:     "test query",
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "valid request with sonar model",
			request: SearchRequest{
				Query:     "test query",
				Model:     "sonar",
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "valid request with sonar-pro model",
			request: SearchRequest{
				Query:     "test query",
				Model:     "sonar-pro",
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "valid request with sonar-reasoning model",
			request: SearchRequest{
				Query:     "test query",
				Model:     "sonar-reasoning",
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "valid request with search mode",
			request: SearchRequest{
				Query:      "test query",
				SearchMode: "academic",
				MaxTokens:  100,
			},
			wantErr: false,
		},
		{
			name: "valid request with date range",
			request: SearchRequest{
				Query:     "test query",
				DateRange: "week",
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "valid request with sources",
			request: SearchRequest{
				Query:     "test query",
				Sources:   []string{"example.com", "test.org"},
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "empty query",
			request: SearchRequest{
				Query:     "",
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidQuery,
		},
		{
			name: "whitespace only query",
			request: SearchRequest{
				Query:     "   ",
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidQuery,
		},
		{
			name: "query exceeds max length",
			request: SearchRequest{
				Query:     string(make([]byte, MaxQueryLength+1)),
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "invalid model",
			request: SearchRequest{
				Query:     "test query",
				Model:     "invalid-model",
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "invalid search mode",
			request: SearchRequest{
				Query:      "test query",
				SearchMode: "invalid-mode",
				MaxTokens:  100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "invalid date range",
			request: SearchRequest{
				Query:     "test query",
				DateRange: "invalid-range",
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "negative max tokens",
			request: SearchRequest{
				Query:     "test query",
				MaxTokens: -1,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "max tokens exceeds limit",
			request: SearchRequest{
				Query:     "test query",
				MaxTokens: 128001,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "zero max tokens allowed",
			request: SearchRequest{
				Query:     "test query",
				MaxTokens: 0,
			},
			wantErr: false,
		},
		{
			name: "too many sources",
			request: SearchRequest{
				Query:     "test query",
				Sources:   make([]string, MaxSourcesCount+1),
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "too many options",
			request: SearchRequest{
				Query: "test query",
				Options: func() map[string]string {
					opts := make(map[string]string)
					for i := 0; i <= MaxOptionsCount; i++ {
						opts[string(rune('a'+i))] = "value"
					}
					return opts
				}(),
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "option key too long",
			request: SearchRequest{
				Query: "test query",
				Options: map[string]string{
					string(make([]byte, MaxOptionKeyLength+1)): "value",
				},
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
		{
			name: "option value too long",
			request: SearchRequest{
				Query: "test query",
				Options: map[string]string{
					"key": string(make([]byte, MaxOptionValueLength+1)),
				},
				MaxTokens: 100,
			},
			wantErr: true,
			errType: ErrInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("SearchRequest.Validate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("SearchRequest.Validate() error = %v, want error type %v", err, tt.errType)
				}
			} else if err != nil {
				t.Errorf("SearchRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}