package miner

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chzyer/readline"
	"github.com/deroproject/derohe/astrobwt/astrobwtv3"
	"github.com/deroproject/derohe/block"
	"github.com/go-logr/logr"
	"github.com/jpillora/backoff"
	"github.com/whalesburg/dero-stratum-miner/internal/config"
	"github.com/whalesburg/dero-stratum-miner/stratum"
)

var reportHashrateInterval = time.Second * 30

type Client struct {
	counter uint64 // Must be the first field. Otherwise atomic operations panic on arm7
	config  *config.Miner
	stratum *stratum.Client
	console *readline.Instance
	logger  logr.Logger

	mu           sync.RWMutex
	job          *stratum.Job
	jobCounter   int64
	iterations   int
	hashrate     uint64
	mining       bool
	miningString string
	diffString   string
	heightString string

	shareCounter    uint64
	acceptedCounter uint64
}

func New(config *config.Miner, stratum *stratum.Client, console *readline.Instance, logger logr.Logger) (*Client, error) {
	c := &Client{
		config:     config,
		stratum:    stratum,
		iterations: 100,
		console:    console,
	}
	c.setLogger(logger)
	return c, nil
}

func (c *Client) Close() error {
	if c.console != nil {
		return c.console.Close()
	}
	return nil
}

func (c *Client) Start(ctx context.Context) error {
	if c.config.Threads < 1 || c.iterations < 1 || c.config.Threads > 2048 {
		panic("Invalid parameters\n")
	}
	if c.config.Threads > 255 {
		c.logger.Error(nil, "This program supports maximum 256 CPU cores.", "available", c.config.Threads)
		c.config.Threads = 255
	}

	go c.gatherStats(ctx)
	if c.config.NonInteractive {
		go c.noniSummary(ctx)
	}

	go c.getwork(ctx)

	for i := 0; i < c.config.Threads; i++ {
		go c.mineblock(ctx, i)
	}

	go c.reportHashrate(ctx)

	if !c.config.NonInteractive {
		c.startConsole(ctx)
	}
	return nil
}

func (c *Client) makeBackoff() backoff.Backoff {
	return backoff.Backoff{
		Min:    time.Second,
		Max:    time.Second * 30,
		Factor: 1.5,
		Jitter: true,
	}
}

func (c *Client) getwork(ctx context.Context) {
	b := c.makeBackoff()
	rand.Seed(time.Now().UTC().UnixNano())

	go func() {
		jobListener := c.stratum.NewJobListener(0)
		defer jobListener.Close()

		for {
			select {
			case j := <-jobListener.Ch():
				c.mu.Lock()
				c.job = j
				c.jobCounter++
				c.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		if err := c.stratum.Dial(ctx); err != nil {
			waitDuration := b.Duration()
			c.logger.Error(err, "Error connecting to server", "server adress", c.config.PoolURL)
			c.logger.Info(fmt.Sprintf("Will try again in %f seconds", waitDuration.Seconds()))
			time.Sleep(waitDuration)
			continue
		}

		//respListener := c.stratum.NewResponseListener(2)
		//go c.listenStratumResponses(respListener)
	}
}

/* func (c *Client) listenStratumResponses(l *broadcast.Listener[*stratum.JobResponse]) {
	defer l.Close()
	for range l.Ch() {
		c.shareCounter = uint64(c.stratum.GetTotalShares())
		c.rejectedCounter = uint64(c.stratum.GetTotalShares() - c.stratum.GetAcceptedShares())
	}
} */

func (c *Client) mineblock(ctx context.Context, tid int) {
	var diff big.Int
	var work [block.MINIBLOCK_SIZE]byte

	var randomBuf [12]byte

	rand.Read(randomBuf[:]) //#nosec G404

	time.Sleep(time.Millisecond * 500)

	nonceBuf := work[block.MINIBLOCK_SIZE-5:] //since slices are linked, it modifies parent
	runtime.LockOSThread()
	threadaffinity()

	var localJobCounter int64

	i := uint32(0)

	for {
		c.mu.RLock()
		myjob := c.job
		localJobCounter = c.jobCounter
		c.mu.RUnlock()
		if myjob == nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}

		n, err := hex.Decode(work[:], []byte(myjob.Blob))
		if err != nil || n != block.MINIBLOCK_SIZE {
			c.logger.Error(err, "Blockwork could not be decoded successfully", "blockwork", myjob.Blob, "n", n, "job", myjob)
			time.Sleep(time.Millisecond * 500)
			continue
		}

		copy(work[block.MINIBLOCK_SIZE-12:], randomBuf[:]) // add more randomization in the mix
		work[block.MINIBLOCK_SIZE-1] = byte(tid)
		diff.SetString(strconv.Itoa(int(myjob.Difficulty)), 10)

		if work[0]&0xf != 1 { // check  version
			c.logger.Error(nil, "Unknown version, please check for updates", "version", work[0]&0x1f)
			time.Sleep(time.Millisecond * 500)
			continue
		}

		for localJobCounter == c.jobCounter { // update job when it comes, expected rate 2 per second
			if !c.stratum.IsConnected() {
				time.Sleep(time.Millisecond * 500)
				continue
			}
			i++
			binary.BigEndian.PutUint32(nonceBuf, i)

			powhash := astrobwtv3.AstroBWTv3(work[:])
			atomic.AddUint64(&c.counter, 1)

			if CheckPowHashBig(powhash, &diff) { // note we are doing a local, NW might have moved meanwhile
				c.logger.V(1).Info("Successfully found share (going to submit)", "difficulty", myjob.Difficulty, "height", myjob.Height)
				func() {
					defer c.recover(1) // nolint: errcheck
					nonce := work[len(work)-12:]
					share := stratum.NewShare(myjob.ID, fmt.Sprintf("%x", nonce), fmt.Sprintf("%x", powhash[:]))
					go func() {
						if err := c.stratum.SubmitShare(ctx, share); err != nil {
							c.logger.Error(err, "Failed to submit share")
						}
					}()
				}()
			}
		}
	}
}

func (c *Client) recover(level int) (err error) {
	if r := recover(); r != nil {
		err = fmt.Errorf("Recovered r:%+v stack %s", r, string(debug.Stack()))
		c.logger.V(level).Error(nil, "Recovered ", "error", r, "stack", string(debug.Stack()))
	}
	return
}

// TODO
func (c *Client) reportHashrate(ctx context.Context) {
	ticker := time.NewTicker(reportHashrateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			/* if err := c.stratum.ReportHashrate(stratum.NewReport(c.GetHashrate())); err != nil {
				c.logger.Error(err, "Failed to report hashrate")
			} */
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) GetHashrate() uint64 {
	return c.hashrate
}

func (c *Client) GetTotalShares() uint64 {
	return c.shareCounter
}

func (c *Client) GetAcceptedShares() uint64 {
	return c.acceptedCounter
}

func (c *Client) GetRejectedShares() uint64 {
	return c.shareCounter - c.acceptedCounter
}

func (c *Client) GetPoolURL() string {
	return c.config.PoolURL
}
