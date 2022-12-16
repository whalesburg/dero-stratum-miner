package stratum

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	stratumrpc "github.com/jon4hz/stratum-rpc"
	"github.com/jpillora/backoff"
	"github.com/teivah/broadcast"
)

var ErrNotConnected = errors.New("stratum: not connected")

const (
	closedState = 1 << iota
	connectingState
	connectedState
	closedForeverState
)

type Client struct {
	state              int
	mu                 sync.RWMutex
	client             *rpc2.Client
	conn               net.Conn
	url                string
	opts               *Opts
	LogFn              logFnOptions
	lastMsg            time.Time
	reconnCancel       context.CancelFunc
	sessionID          string
	jobBroadcaster     *broadcast.Relay[*Job]
	respBroadcaster    *broadcast.Relay[*JobResponse]
	submittedJobsIdsMu sync.Mutex
	lastSubmittedShare *Share
}

type logFnOptions struct {
	Debug func(string)
	Info  func(string)
	Error func(error, string)
}

func New(url string, opts ...OptsFunc) *Client {
	c := &Client{
		url: url,
		opts: &Opts{
			keepaliveTimeout:        time.Second * 25,
			reconnectIntervalMin:    1 * time.Second,
			reconnectIntervalMax:    30 * time.Second,
			reconnectIntervalFactor: 1.5,
			readTimeout:             time.Second * 15,
			writeTimeout:            time.Second * 15,
		},
		LogFn: logFnOptions{
			Debug: func(x string) {
				log.Printf("[DEBU] %s", x)
			},
			Info: func(x string) {
				log.Printf("[INFO] %s", x)
			},
			Error: func(err error, x string) {
				log.Printf("[ERRO] %s: %s", x, err.Error())
			},
		},
		jobBroadcaster:     broadcast.NewRelay[*Job](),
		lastSubmittedShare: &Share{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) makeBackoff() backoff.Backoff {
	return backoff.Backoff{
		Min:    c.opts.reconnectIntervalMin,
		Max:    c.opts.reconnectIntervalMax,
		Factor: c.opts.reconnectIntervalFactor,
		Jitter: true,
	}
}

func (c *Client) isState(s int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return (c.state & s) > 0
}

func (c *Client) setStateIfNot(targetState, conditionState int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.state&conditionState == 0 {
		c.state = targetState
		return true
	}
	return false
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == connectedState
}

func (c *Client) Dial(ctx context.Context) error {
	if !c.setStateIfNot(connectingState, connectingState|connectedState) {
		return nil
	}

	if err := c.dial(ctx); err != nil {
		return err
	}
	go c.checkLastMsg(ctx)
	return nil
}

func (c *Client) dial(ctx context.Context) error {
	var err error
	d := net.Dialer{KeepAlive: c.opts.keepaliveTimeout}
	c.mu.Lock()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(c.opts.writeTimeout))
	defer cancel()

	if c.opts.useTLS {
		td := tls.Dialer{NetDialer: &d, Config: &tls.Config{
			MinVersion:         tls.VersionTLS13,
			InsecureSkipVerify: c.opts.ignoreTLSValidation, //nolint:gosec
		}}
		c.conn, err = td.DialContext(ctx, "tcp", c.url)
	} else {
		c.conn, err = d.DialContext(ctx, "tcp", c.url)
	}
	if err == nil {
		c.state = connectedState
	} else {
		c.state = closedState
	}
	c.mu.Unlock()
	if err != nil {
		return err
	}
	c.LogFn.Info("successfully connected to pool: " + c.url)

	c.client = rpc2.NewClientWithCodec(stratumrpc.NewStratumCodec(c.conn))
	c.client.Handle("job", c.HandleJob)
	go c.client.Run()

	var params map[string]any
	msg := MiningAuthorizeRequest{c.opts.username, c.opts.password, c.opts.agentName}
	params, err = msg.Encode()
	if err != nil {
		return err
	}

	var res any
	err = c.client.CallWithContext(ctx, "login", params, &res)
	if err != nil {
		return err
	}

	sid, ok := res.(map[string]any)
	if !ok {
		return ErrNoSessionID
	}

	_, ok = sid["id"].(string)
	if !ok {
		return ErrNoSessionID
	}

	data, ok := sid["job"].(map[string]any)
	if !ok {
		return fmt.Errorf("Failed to cast job")
	}

	job := &Job{}
	if err := job.Decode(data); err != nil {
		return err
	}
	c.broadcastJob(job)

	return nil
}

func (c *Client) CloseAndReconnect(ctx context.Context) {
	c.Close(false)
	go c.reconnect(ctx)
}

func (c *Client) Close(forever bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state&(closedState|closedForeverState) > 0 {
		return
	}

	if c.conn != nil {
		c.client.Close()
		if err := c.conn.Close(); err != nil {
			c.LogFn.Error(err, "connection closing error")
		}
	}

	if forever {
		c.state = closedForeverState
	} else {
		c.state = closedState
	}
}

func (c *Client) Shutdown() {
	if c.isState(closedState | closedForeverState) {
		return
	}
	c.Close(true)
}

func (c *Client) checkLastMsg(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if time.Since(c.lastMsg) > time.Minute*3 {
				c.LogFn.Error(errors.New("no messages for 3 minutes"), "dead connection?, reconnecting...")
				c.CloseAndReconnect(ctx)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) reconnect(ctx context.Context) {
	if !c.setStateIfNot(connectingState, connectingState|connectedState|closedForeverState) {
		return
	}

	b := c.makeBackoff()
	rand.Seed(time.Now().UTC().UnixNano())

	if c.reconnCancel != nil {
		c.reconnCancel()
	}
	reconnCtx, cancel := context.WithCancel(ctx)
	c.reconnCancel = cancel

	for {
		select {
		case <-ctx.Done():
			c.reconnCancel()
			return
		case <-reconnCtx.Done():
			c.LogFn.Debug("reconnect cancelled")
			return
		default:
			err := c.dial(reconnCtx)
			if err == nil {
				return
			}

			waitDuration := b.Duration()
			c.LogFn.Error(err, fmt.Sprintf("dial error, will try again in %f seconds", waitDuration.Seconds()))
			time.Sleep(waitDuration)
		}
	}
}

func (c *Client) broadcastJob(job *Job) {
	c.LogFn.Debug(fmt.Sprintf("received job %s", job.ID))
	c.jobBroadcaster.Notify(job)
}

func (c *Client) NewJobListener(buff int) *broadcast.Listener[*Job] {
	c.LogFn.Debug(fmt.Sprintf("registered job listener, buff: %d", buff))
	return c.jobBroadcaster.Listener(buff)
}

func (c *Client) NewResponseListener(buff int) *broadcast.Listener[*JobResponse] {
	c.LogFn.Debug(fmt.Sprintf("registered response listener, buff: %d", buff))
	return c.respBroadcaster.Listener(buff)
}
