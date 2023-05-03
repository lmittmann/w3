package w3

import "golang.org/x/time/rate"

// A ClientOption configures a Client.
type ClientOption func(*Client)

func WithRateLimiter(rl *rate.Limiter) ClientOption {
	return func(c *Client) {
		c.rl = rl
	}
}
