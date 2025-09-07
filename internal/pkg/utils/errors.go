package utils

import (
	"errors"
	"fmt"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

func WrapError(err error, domainErr error, message string, args ...any) error {
	if err == nil {
		return nil
	}

	wrappedMsg := fmt.Sprintf(message, args...)
	return fmt.Errorf("%s: %w: %w", wrappedMsg, domainErr, err)
}

func IsAPIError(err error) bool {
	return err != nil && (errors.Is(err, domain.ErrAPIError) || errors.Is(err, domain.ErrRateLimited))
}

func IsNetworkError(err error) bool {
	return err != nil && errors.Is(err, domain.ErrNetworkError)
}

func IsTimeoutError(err error) bool {
	return err != nil && errors.Is(err, domain.ErrTimeout)
}
