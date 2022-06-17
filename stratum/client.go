package stratum

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/teivah/broadcast"
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

type Client struct {
	id               int
	url              string
	username         string
	password         string
	keepaliveTimeout time.Duration
	useTLS           bool
	parentCtx        context.Context
	ctx              context.Context
	cancel           context.CancelFunc
	conn             net.Conn
	reader           *bufio.Reader
	connected        bool
	sessionID        string
	jobBroadcaster   *broadcast.Relay[*Job]
	respBroadcaster  *broadcast.Relay[*Response]

	submittedShares int
	acceptedShares  int

	submittedJobIds    map[int]struct{}
	submittedJobsIdsMu sync.Mutex
	lastSubmittedShare *Share

	submitMu sync.Mutex
}

func New(url string, opts ...Opts) *Client {
	c := &Client{
		url:                url,
		parentCtx:          context.Background(),
		keepaliveTimeout:   time.Second * 15,
		jobBroadcaster:     broadcast.NewRelay[*Job](),
		respBroadcaster:    broadcast.NewRelay[*Response](),
		lastSubmittedShare: &Share{},
		submittedJobIds:    make(map[int]struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	c.ctx, c.cancel = context.WithCancel(c.parentCtx)
	return c
}

func (c *Client) NewJobListener(buff int) *broadcast.Listener[*Job] {
	return c.jobBroadcaster.Listener(buff)
}

func (c *Client) NewResponseListener(buff int) *broadcast.Listener[*Response] {
	return c.respBroadcaster.Listener(buff)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Connect() error {
	var err error
	d := net.Dialer{KeepAlive: c.keepaliveTimeout}
	if c.useTLS {
		td := tls.Dialer{NetDialer: &d, Config: &tls.Config{
			MinVersion: tls.VersionTLS13,
		}}
		c.conn, err = td.DialContext(c.ctx, "tcp", c.url)
	} else {
		c.conn, err = d.DialContext(c.ctx, "tcp", c.url)
	}
	if err != nil {
		return err
	}
	c.reader = bufio.NewReader(c.conn)
	return nil
}

func (c *Client) call(method string, args any) (*Request, error) {
	c.submitMu.Lock()
	defer c.submitMu.Unlock()
	c.id++
	req := NewRequest(c.id, method, args)
	data, err := req.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %v", err)
	}

	//fmt.Println("Sending:", string(data))

	if _, err := c.conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	return req, nil
}

func (c *Client) handleMessages() {
	// This loop only ends on error
	defer func() {
		//sc.Reconnect()
	}()

	for {
		line, err := c.readLine()
		if err != nil {
			//TODO: debug logger
			break
		}
		//TODO: debug logger
		var msg map[string]interface{}
		if err = json.Unmarshal(line, &msg); err != nil {
			//TODO: debug logger
			break
		}

		id := msg["id"]
		switch id.(type) {
		case uint64, float64:
			// This is a response
			response, err := parseResponse(line)
			if err != nil {
				//TODO: debug logger
				continue
			}
			isError := false
			if response.Result == nil {
				// This is an error
				isError = true
			}
			id := int(response.ID.(float64))

			c.submittedJobsIdsMu.Lock()
			if _, ok := c.submittedJobIds[id]; ok {
				if !isError {
					// This is a response from the server signalling that our work has been accepted
					delete(c.submittedJobIds, id)
					c.acceptedShares++
					c.submittedShares++
					//TODO: debug logger
				} else {
					delete(c.submittedJobIds, id)
					c.submittedShares++
					//TODO: debug logger
				}
			} else {
				statusIntf, ok := response.Result["status"]
				if !ok {
					//TODO: debug logger
				} else {
					status := statusIntf.(string)
					switch status {
					case "OK":
						//TODO: debug logger
					}
				}
			}
			c.submittedJobsIdsMu.Unlock()
			c.respBroadcaster.Notify(response)

		default:
			// this is a notification
			//TODO: debug logger
			switch msg["method"].(string) {
			case "job":
				if job, err := extractJob(msg["params"].(map[string]interface{})); err != nil {
					//TODO: debug logger
					continue
				} else {
					c.jobBroadcaster.Notify(job)
				}
			default:
				//TODO: debug logger
			}
		}
	}
}

func (c *Client) GetTotalShares() int {
	return c.submittedShares
}

func (c *Client) GetAcceptedShares() int {
	return c.acceptedShares
}
