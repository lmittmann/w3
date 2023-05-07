package w3

import "golang.org/x/time/rate"

// An Option configures a Client.
type Option func(*Client)

// WithRateLimiter sets the rate limiter for the client. If perCall is true, the
// rate limiter is applied to each call. Otherwise, the rate limiter is applied
// to each request.
func WithRateLimiter(rl *rate.Limiter, perCall bool) Option {
	return func(c *Client) {
		c.rl = rl
		c.rlPerCall = perCall
	}
}
