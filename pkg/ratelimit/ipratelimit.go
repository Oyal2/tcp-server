package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type IP struct {
	Count     int
	LastReset time.Time
}

type IPRateLimiter struct {
	mu       sync.RWMutex
	ips      map[string]*IP
	limit    int
	interval time.Duration
}

func NewIPRateLimiter(limit int, interval time.Duration) (*IPRateLimiter, error) {
	// the IP rate limit should be greater than 0
	if limit < 1 {
		return nil, fmt.Errorf("limit cannot be %d", limit)
	}

	rl := IPRateLimiter{
		ips:      make(map[string]*IP),
		limit:    limit,
		interval: interval,
	}
	return &rl, nil
}

func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// check if we had cached the ip
	entry, exists := rl.ips[ip]

	// if the ip doesnt exists then lets cache it and count it
	if !exists {
		rl.ips[ip] = &IP{Count: 1, LastReset: now}
		return true
	}

	// if the last time we reset the limit is passed the intervalled time, then we will reset.
	if now.Sub(entry.LastReset) > rl.interval {
		entry.Count = 1
		entry.LastReset = now
		return true
	}

	// if the limit is larger than what the ip's count is, then we will deny access
	if entry.Count >= rl.limit {
		return false
	}

	// increment the count
	entry.Count++
	return true
}

func (rl *IPRateLimiter) Clean() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Go through all the ips in the system and delete anything that is old
	for ip, entry := range rl.ips {
		if time.Since(entry.LastReset) > rl.interval {
			delete(rl.ips, ip)
		}
	}
}

func (rl *IPRateLimiter) IPs() map[string]*IP {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.ips
}

func (rl *IPRateLimiter) Interval() time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.interval
}
