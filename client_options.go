package w3

import "golang.org/x/time/rate"

// A ClientOption configures a Client.
type ClientOption func(*Client)

// WithRateLimiter sets the rate limiter for the client.
func WithRateLimiter(rl *rate.Limiter) ClientOption {
	return func(c *Client) {
		c.rl = rl
	}
}
