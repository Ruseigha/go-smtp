package retry

import (
	"context"
	"fmt"
	"math"
	"time"
)

type Config struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

type RetryableFunc func(ctx context.Context) error

func Do(ctx context.Context, config Config, fn RetryableFunc) error {
	var lastErr error
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Try the function
		err := fn(ctx)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if we should retry
		if !IsRetryable(err) {
			return fmt.Errorf("permanent error (attempt %d/%d): %w", attempt, config.MaxAttempts, err)
		}
		
		// Last attempt, don't wait
		if attempt == config.MaxAttempts {
			break
		}
		
		// Calculate backoff delay
		delay := calculateBackoff(attempt, config)
		
		fmt.Printf("Attempt %d/%d failed: %v. Retrying in %v...\n", attempt, config.MaxAttempts, err, delay)
		
		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}
	
	return fmt.Errorf("max attempts reached (%d): %w", config.MaxAttempts, lastErr)
}

func calculateBackoff(attempt int, config Config) time.Duration {
	// Exponential backoff: delay = initialDelay * (multiplier ^ (attempt - 1))
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt-1))
	
	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	return time.Duration(delay)
}

// IsRetryable determines if an error is temporary and worth retrying
func IsRetryable(err error) bool {
	// Check error message for temporary indicators
	errStr := err.Error()
	
	// Temporary SMTP errors (4xx)
	temporaryErrors := []string{
		"421", // Service not available
		"450", // Mailbox unavailable
		"451", // Local error
		"452", // Insufficient storage
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
	}
	
	for _, tempErr := range temporaryErrors {
		if contains(errStr, tempErr) {
			return true
		}
	}
	
	// Permanent errors (5xx) - don't retry
	permanentErrors := []string{
		"550", // Mailbox unavailable (user doesn't exist)
		"551", // User not local
		"552", // Exceeded storage
		"553", // Mailbox name invalid
		"554", // Transaction failed
	}
	
	for _, permErr := range permanentErrors {
		if contains(errStr, permErr) {
			return false
		}
	}
	
	// Default: retry for unknown errors
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}