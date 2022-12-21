package utils

import "time"

// RetryStopper is an error type that can be used to stop retries.
type RetryStopper struct {
	error
}

// Retry calls the given function and retries it if an error is returned.
// The number of attempts and the sleep duration between attempts can be specified.
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(RetryStopper); ok {
			// Return the original error for later checking.
			return s.error
		}

		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

// NewRetryStopper creates a new RetryStopper with the given error.
func NewRetryStopper(err error) RetryStopper {
	return RetryStopper{err}
}
