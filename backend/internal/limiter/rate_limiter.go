package limiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	Tokens         float64   //Current number of tokens
	MaxTokens      float64   //Max number of tokens
	RefillRate     float64   //Tokens added per second
	LastRefillTime time.Time //Last time tokens were refilled
	Mutex          sync.Mutex
}

// Maps IP addresses to their respective rate limiters
type IPRateLimiter struct {
	Limiters map[string]*RateLimiter
	Mutex    sync.Mutex
}

// Creates new instance of a RateLimiter
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		Tokens:         maxTokens,
		MaxTokens:      maxTokens,
		RefillRate:     refillRate,
		LastRefillTime: time.Now(),
	}
}

func NewIPRateLimiter() *IPRateLimiter {
	return &IPRateLimiter{
		Limiters: make(map[string]*RateLimiter),
	}
}

// Gets limiter for IP address from struct map
func (i *IPRateLimiter) GetLimiter(ip string, limit, refresh float64) *RateLimiter {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()

	limiter, exists := i.Limiters[ip]
	if !exists {
		limiter = NewRateLimiter(limit, refresh)
		i.Limiters[ip] = limiter
	}

	return limiter
}

// Refill available tokens in limiter struct
func (r *RateLimiter) RefillTokens() {

	now := time.Now()
	duration := now.Sub(r.LastRefillTime).Seconds()
	tokensToAdd := duration * r.RefillRate

	r.Tokens += tokensToAdd
	if r.Tokens > r.MaxTokens {
		r.Tokens = r.MaxTokens
	}
	r.LastRefillTime = now
}

// Determines whether a request will be allowed or rejected
func (r *RateLimiter) Allow() bool {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.RefillTokens()

	if r.Tokens >= 1 {
		r.Tokens--
		return true
	}

	return false
}
