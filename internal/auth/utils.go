package auth

import "time"

func ParseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
