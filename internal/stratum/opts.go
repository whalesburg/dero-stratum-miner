package stratum

import (
	"context"
	"time"
)

type Opts func(*Client)

func WithUsername(username string) Opts {
	return func(c *Client) {
		c.username = username
	}
}

func WithPassword(password string) Opts {
	return func(c *Client) {
		c.password = password
	}
}

func WithReadTimeout(timeout time.Duration) Opts {
	return func(c *Client) {
		c.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Opts {
	return func(c *Client) {
		c.writeTimeout = timeout
	}
}

func WithKeepaliveTimeout(timeout time.Duration) Opts {
	return func(c *Client) {
		c.keepaliveTimeout = timeout
	}
}

func WithContext(ctx context.Context) Opts {
	return func(c *Client) {
		c.parentCtx = ctx
	}
}

func WithUseTLS() Opts {
	return func(c *Client) {
		c.useTLS = true
	}
}

func WithDebugLogger(logger func(string)) Opts {
	return func(c *Client) {
		c.LogFn.Debug = logger
	}
}

func WithInfoLogger(logger func(string)) Opts {
	return func(c *Client) {
		c.LogFn.Info = logger
	}
}

func WithErrorLogger(logger func(error, string)) Opts {
	return func(c *Client) {
		c.LogFn.Error = logger
	}
}

func WithReconnectIntervalMin(interval time.Duration) Opts {
	return func(c *Client) {
		c.reconnectIntervalMin = interval
	}
}

func WithReconnectIntervalMax(interval time.Duration) Opts {
	return func(c *Client) {
		c.reconnectIntervalMax = interval
	}
}

func WithReconnectIntervalFactor(factor float64) Opts {
	return func(c *Client) {
		c.reconnectIntervalFactor = factor
	}
}

func WithAgentName(agentName string) Opts {
	return func(c *Client) {
		c.agentName = agentName
	}
}
