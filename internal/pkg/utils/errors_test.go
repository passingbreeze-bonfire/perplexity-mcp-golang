package utils

import (
	"fmt"
	"testing"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		domainErr error
		message   string
		args      []any
		wantNil   bool
	}{
		{
			name:      "nil error returns nil",
			err:       nil,
			domainErr: domain.ErrAPIError,
			message:   "test message",
			wantNil:   true,
		},
		{
			name:      "wraps error correctly",
			err:       fmt.Errorf("original error"),
			domainErr: domain.ErrAPIError,
			message:   "API call failed",
			wantNil:   false,
		},
		{
			name:      "formats message with args",
			err:       fmt.Errorf("original error"),
			domainErr: domain.ErrNetworkError,
			message:   "network call to %s failed",
			args:      []any{"example.com"},
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.err, tt.domainErr, tt.message, tt.args...)
			if tt.wantNil {
				if result != nil {
					t.Errorf("WrapError() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("WrapError() = nil, want non-nil error")
				}
			}
		})
	}
}

func TestIsAPIError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "API error",
			err:  domain.ErrAPIError,
			want: true,
		},
		{
			name: "rate limited error",
			err:  domain.ErrRateLimited,
			want: true,
		},
		{
			name: "wrapped API error",
			err:  fmt.Errorf("wrapper: %w", domain.ErrAPIError),
			want: true,
		},
		{
			name: "wrapped rate limited error",
			err:  fmt.Errorf("wrapper: %w", domain.ErrRateLimited),
			want: true,
		},
		{
			name: "network error",
			err:  domain.ErrNetworkError,
			want: false,
		},
		{
			name: "other error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAPIError(tt.err); got != tt.want {
				t.Errorf("IsAPIError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "network error",
			err:  domain.ErrNetworkError,
			want: true,
		},
		{
			name: "wrapped network error",
			err:  fmt.Errorf("wrapper: %w", domain.ErrNetworkError),
			want: true,
		},
		{
			name: "API error",
			err:  domain.ErrAPIError,
			want: false,
		},
		{
			name: "other error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNetworkError(tt.err); got != tt.want {
				t.Errorf("IsNetworkError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "timeout error",
			err:  domain.ErrTimeout,
			want: true,
		},
		{
			name: "wrapped timeout error",
			err:  fmt.Errorf("wrapper: %w", domain.ErrTimeout),
			want: true,
		},
		{
			name: "API error",
			err:  domain.ErrAPIError,
			want: false,
		},
		{
			name: "other error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeoutError(tt.err); got != tt.want {
				t.Errorf("IsTimeoutError() = %v, want %v", got, tt.want)
			}
		})
	}
}
