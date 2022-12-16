package stratum

import (
	"github.com/cenkalti/rpc2"
)

func (c *Client) HandleJob(client *rpc2.Client, params map[string]any, res *any) error {
	job := &Job{}
	if err := job.Decode(params); err != nil {
		return err
	}
	c.broadcastJob(job)
	return nil
}

type JobResponse struct{}
