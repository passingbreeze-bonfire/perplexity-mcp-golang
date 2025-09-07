package utils

import (
	"context"
	"time"
)

func EnsureTimeout(ctx context.Context, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		return context.WithTimeout(ctx, defaultTimeout)
	}

	return ctx, func() {}
}

func WithSafetyTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

func IsCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
