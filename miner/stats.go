package miner

import (
	"context"
	"fmt"
	"time"

	"github.com/jon4hz/hashconv"
)

func (c *Client) gatherStats(ctx context.Context) {
	var (
		lastCounter     = uint64(0)
		lastCounterTime = time.Now()
		lastUpdate      = time.Now()
		miningString    string
		diffString      string
		heightString    string
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// we assume that the miner stopped if the conolse wasn't updated within the last five seconds.
		if time.Since(lastUpdate) > time.Second*5 {
			if c.mining {
				miningString = "\033[31mNot Mining"
				testnetString := ""
				if c.config.Testnet {
					testnetString = "\033[31m Testnet"
				}
				c.setPrompt(heightString, diffString, miningString, testnetString)
				c.mining = false
			}
		} else {
			c.mining = true
		}

		// only update prompt if needed
		if lastCounter != c.counter {
			if c.mining {
				c.heightString = fmt.Sprintf("%.0f", c.job.Height)
				heightString = fmt.Sprintf("\033[33mHeight %s", c.heightString)

				switch {
				case c.job.Difficulty > 1_000_000_000:
					c.diffString = fmt.Sprintf("%.1fG", float32(c.job.Difficulty)/1_000_000_000.0)
					diffString = fmt.Sprintf("\033[32mDiff %s", c.diffString)
				case c.job.Difficulty > 1_000_000:
					c.diffString = fmt.Sprintf("%.1fM", float32(c.job.Difficulty)/1_000_000.0)
					diffString = fmt.Sprintf("\033[32mDiff %s", c.diffString)
				case c.job.Difficulty > 1000:
					c.diffString = fmt.Sprintf("%.1fK", float32(c.job.Difficulty)/1000.0)
					diffString = fmt.Sprintf("\033[32mDiff %s", c.diffString)
				case c.job.Difficulty > 0:
					c.diffString = fmt.Sprintf("%d", c.job.Difficulty)
					diffString = fmt.Sprintf("\033[32mDiff %s", c.diffString)
				}

				miningSpeed := float64(c.counter-lastCounter) / (float64(uint64(time.Since(lastCounterTime))) / 1_000_000_000.0)
				c.hashrate = uint64(miningSpeed)
				lastCounter = c.counter
				lastCounterTime = time.Now()
				c.miningString = fmt.Sprintf("%s/s", hashconv.Format(int64(miningSpeed)))
				miningString = fmt.Sprintf("Mining @ %s", c.miningString)
			}

			testnetString := ""
			if c.config.Testnet {
				testnetString = "\033[31m Testnet"
			}

			c.setPrompt(heightString, diffString, miningString, testnetString)
			lastUpdate = time.Now()
		}
		time.Sleep(1 * time.Second)
	}
}
func (c *Client) noniSummary(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			c.printSummary()
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) printSummary() {
	c.logger.Info("Summary",
		"height", c.heightString,
		"diff", c.diffString,
		"accepted", c.GetAcceptedShares(),
		"rejected", c.GetRejectedShares(),
		"hashrate", c.miningString,
	)
}
