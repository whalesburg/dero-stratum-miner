package stratum

import "time"

type Opts struct {
	agentName               string
	username                string
	password                string
	readTimeout             time.Duration
	writeTimeout            time.Duration
	keepaliveTimeout        time.Duration
	reconnectIntervalMin    time.Duration
	reconnectIntervalMax    time.Duration
	reconnectIntervalFactor float64
	useTLS                  bool
	ignoreTLSValidation     bool
}

type OptsFunc func(*Client)

func WithUsername(username string) OptsFunc {
	return func(c *Client) {
		c.opts.username = username
	}
}

func WithPassword(password string) OptsFunc {
	return func(c *Client) {
		c.opts.password = password
	}
}

func WithReadTimeout(timeout time.Duration) OptsFunc {
	return func(c *Client) {
		c.opts.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) OptsFunc {
	return func(c *Client) {
		c.opts.writeTimeout = timeout
	}
}

func WithKeepaliveTimeout(timeout time.Duration) OptsFunc {
	return func(c *Client) {
		c.opts.keepaliveTimeout = timeout
	}
}
func WithUseTLS() OptsFunc {
	return func(c *Client) {
		c.opts.useTLS = true
	}
}

func WithIgnoreTLSValidation(ignoreTLSValidation bool) OptsFunc {
	return func(c *Client) {
		c.opts.ignoreTLSValidation = ignoreTLSValidation
	}
}
func WithDebugLogger(logger func(string)) OptsFunc {
	return func(c *Client) {
		c.LogFn.Debug = logger
	}
}

func WithInfoLogger(logger func(string)) OptsFunc {
	return func(c *Client) {
		c.LogFn.Info = logger
	}
}

func WithErrorLogger(logger func(error, string)) OptsFunc {
	return func(c *Client) {
		c.LogFn.Error = logger
	}
}

func WithReconnectIntervalMin(interval time.Duration) OptsFunc {
	return func(c *Client) {
		c.opts.reconnectIntervalMin = interval
	}
}

func WithReconnectIntervalMax(interval time.Duration) OptsFunc {
	return func(c *Client) {
		c.opts.reconnectIntervalMax = interval
	}
}

func WithReconnectIntervalFactor(factor float64) OptsFunc {
	return func(c *Client) {
		c.opts.reconnectIntervalFactor = factor
	}
}

func WithAgentName(agentName string) OptsFunc {
	return func(c *Client) {
		c.opts.agentName = agentName
	}
}
