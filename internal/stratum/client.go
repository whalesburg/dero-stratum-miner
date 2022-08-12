package stratum

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

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
	state                   int
	mu                      sync.RWMutex
	id                      int
	url                     string
	username                string
	password                string
	readTimeout             time.Duration
	writeTimeout            time.Duration
	keepaliveTimeout        time.Duration
	reconnectIntervalMin    time.Duration
	reconnectIntervalMax    time.Duration
	reconnectIntervalFactor float64
	useTLS                  bool
	parentCtx               context.Context
	ctx                     context.Context
	cancel                  context.CancelFunc
	conn                    net.Conn
	reader                  *bufio.Reader
	connected               bool
	sessionID               string
	jobBroadcaster          *broadcast.Relay[*Job]
	respBroadcaster         *broadcast.Relay[*Response]

	submittedShares int
	acceptedShares  int

	submittedJobIds    map[int]struct{}
	submittedJobsIdsMu sync.Mutex
	lastSubmittedShare *Share

	submitMu sync.Mutex
	LogFn    logFnOptions

	msgHandlerCtx    context.Context
	msgHandlerCancel context.CancelFunc
}

type logFnOptions struct {
	Debug func(string)
	Info  func(string)
	Error func(error, string)
}

func New(url string, opts ...Opts) *Client {
	c := &Client{
		url:                     url,
		parentCtx:               context.Background(),
		keepaliveTimeout:        time.Second * 15,
		reconnectIntervalMin:    1 * time.Second,
		reconnectIntervalMax:    30 * time.Second,
		reconnectIntervalFactor: 1.5,
		jobBroadcaster:          broadcast.NewRelay[*Job](),
		respBroadcaster:         broadcast.NewRelay[*Response](),
		lastSubmittedShare:      &Share{},
		submittedJobIds:         make(map[int]struct{}),
		LogFn: logFnOptions{
			Debug: func(string) {},
			Info:  func(string) {},
			Error: func(error, string) {},
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.ctx, c.cancel = context.WithCancel(c.parentCtx)
	return c
}

func (c *Client) CloseAndReconnect() {
	c.Close(false)
	go func() {
		if err := c.reconnect(); err != nil {
			c.LogFn.Error(err, "connection error")
		}
	}()
}

func (c *Client) Close(forever bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.msgHandlerCancel != nil {
		c.msgHandlerCancel()
	}

	if c.state&(closedState|closedForeverState) > 0 {
		return
	}
	if err := c.conn.Close(); err != nil {
		c.LogFn.Error(err, "connection closing error")
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

func (c *Client) NewJobListener(buff int) *broadcast.Listener[*Job] {
	c.LogFn.Debug(fmt.Sprintf("registered job listener, buff: %d", buff))
	return c.jobBroadcaster.Listener(buff)
}

func (c *Client) NewResponseListener(buff int) *broadcast.Listener[*Response] {
	c.LogFn.Debug(fmt.Sprintf("registered response listener, buff: %d", buff))
	return c.respBroadcaster.Listener(buff)
}

func (c *Client) Dial() error {
	if !c.setStateIfNot(connectingState, connectingState|connectedState) {
		return nil
	}

	if err := c.dial(); err != nil {
		return err
	}
	return nil
}

func (c *Client) dial() error {
	var err error
	d := net.Dialer{KeepAlive: c.keepaliveTimeout}
	c.mu.Lock()
	if c.useTLS {
		td := tls.Dialer{NetDialer: &d, Config: &tls.Config{
			MinVersion: tls.VersionTLS13,
		}}
		c.conn, err = td.DialContext(c.ctx, "tcp", c.url)
	} else {
		c.conn, err = d.DialContext(c.ctx, "tcp", c.url)
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
	c.reader = bufio.NewReader(c.conn)

	if er := c.authorize(); er != nil {
		return errors.New("authorization error: " + er.Error())
	}
	c.handleMessages()
	return nil
}

func (c *Client) reconnect() error {
	if !c.setStateIfNot(connectingState, connectingState|connectedState|closedForeverState) {
		return nil
	}

	b := c.makeBackoff()
	rand.Seed(time.Now().UTC().UnixNano())

	for {
		err := c.dial()
		if err == nil {
			return nil
		}

		waitDuration := b.Duration()
		c.LogFn.Error(err, fmt.Sprintf("dial error, will try again in %f seconds", waitDuration.Seconds()))
		time.Sleep(waitDuration)
	}
}

func (c *Client) makeBackoff() backoff.Backoff {
	return backoff.Backoff{
		Min:    c.reconnectIntervalMin,
		Max:    c.reconnectIntervalMax,
		Factor: c.reconnectIntervalFactor,
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

func (c *Client) call(method string, args any) (*Request, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}
	c.submitMu.Lock()
	defer c.submitMu.Unlock()

	c.id++
	req := NewRequest(c.id, method, args)
	data, err := req.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %v", err)
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)) // nolint: errcheck
	if _, err := c.conn.Write(data); err != nil {
		c.CloseAndReconnect()
		c.LogFn.Error(err, "failed to write request")
		return nil, err
	}
	return req, nil
}

func (c *Client) handleMessages() {
	c.mu.Lock()
	if c.msgHandlerCancel != nil {
		c.msgHandlerCancel()
	}
	c.msgHandlerCtx, c.msgHandlerCancel = context.WithCancel(context.Background())
	c.mu.Unlock()
	go func() {
		for {
			select {
			case <-c.msgHandlerCtx.Done():
				return
			default:
			}

			line, err := c.readLine()
			if err != nil {
				c.LogFn.Error(err, "failed to read line")
				break
			}

			// MAYBE: debug logger
			var msg map[string]interface{}
			if err = json.Unmarshal(line, &msg); err != nil {
				c.LogFn.Error(err, "failed to unmarshal message")
				continue
			}

			id := msg["id"]
			switch id.(type) {
			case uint64, float64:
				// This is a response
				response, err := parseResponse(line)
				if err != nil {
					c.LogFn.Error(err, "failed to parse response")
					continue
				}
				isError := false
				if response.Result == nil {
					// This is an error
					isError = true
				} else {
					isError = response.Error != nil
				}
				id := int(response.ID.(float64))

				c.submittedJobsIdsMu.Lock()
				if _, ok := c.submittedJobIds[id]; ok {
					if !isError {
						// This is a response from the server signalling that our work has been accepted
						delete(c.submittedJobIds, id)
						c.acceptedShares++
						c.submittedShares++
						c.LogFn.Info("accepted share")
					} else {
						delete(c.submittedJobIds, id)
						c.submittedShares++
						c.LogFn.Info("rejected share")
					}
				} else {
					statusIntf, ok := response.Result.(map[string]any)
					if ok {
						s, ok := statusIntf["status"]
						if !ok {
							c.LogFn.Error(errors.New("invalid response"), fmt.Sprintf("failed to parse result: %v", response.Result))
							continue
						}
						status := s.(string)
						switch status {
						case "OK":
							// MAYBE: debug logger
						}
					}
				}
				c.submittedJobsIdsMu.Unlock()
				c.respBroadcaster.Notify(response)

			default:
				// this is a notification
				// MAYBE: debug logger
				switch msg["method"].(string) {
				case "job":
					if job, err := extractJob(msg["params"].(map[string]interface{})); err != nil {
						c.LogFn.Error(err, "failed to extract job")
						continue
					} else {
						c.broadcastJob(job)
					}
				default:
					// MAYBE: debug logger
				}
			}
		}
	}()
}

func (c *Client) GetTotalShares() int {
	return c.submittedShares
}

func (c *Client) GetAcceptedShares() int {
	return c.acceptedShares
}

func (c *Client) readLine() ([]byte, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}
	c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)) // nolint: errcheck
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		c.CloseAndReconnect()
		return nil, err
	}
	return line, nil
}

func parseResponse(b []byte) (*Response, error) {
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
