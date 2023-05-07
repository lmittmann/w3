package w3

import "golang.org/x/time/rate"

// A ClientOption configures a Client.
type ClientOption func(*Client)

// WithRateLimiter sets the rate limiter for the client. If perCall is true, the
// rate limiter is applied to each call. Otherwise, the rate limiter is applied
// to each request.
func WithRateLimiter(rl *rate.Limiter, perCall bool) ClientOption {
	return func(c *Client) {
		c.rl = rl
		c.rlPerCall = perCall
	}
}
