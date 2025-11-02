package utils

import "time"

func RetryBackoff(attempts int) time.Duration {
	if attempts <= 0 {
		attempts = 1
	}
	base := time.Second
	d := base * time.Duration(1<<(uint(attempts-1)))
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}
