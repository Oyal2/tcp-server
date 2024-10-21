package ratelimit

type RateLimiter interface {
	Allow(ip string) bool
	Clean()
}
