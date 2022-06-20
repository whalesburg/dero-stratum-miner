package stratum

import (
	"fmt"
)

type Share struct {
	ID     string `json:"id"`
	JobID  string `json:"job_id"`
	Nonce  string `json:"nonce"`
	Result string `json:"result"`
}

func NewShare(jobID string, nonce string, result string) *Share {
	return &Share{
		ID:     "",
		JobID:  jobID,
		Nonce:  nonce,
		Result: result,
	}
}

func (c *Client) SubmitShare(s *Share) error {
	if s.Result == c.lastSubmittedShare.Result { // should we do this?
		// TODO: debug logger
		return nil
	}

	args := make(map[string]interface{})
	args["id"] = c.sessionID
	args["job_id"] = s.JobID
	args["result"] = s.Result
	args["nonce"] = s.Nonce
	req, err := c.call("submit", args)
	if err != nil {
		return err
	}
	id, ok := req.ID.(int)
	if !ok {
		return fmt.Errorf("failed to convert id to int: %v", req.ID)
	}
	c.submittedJobsIdsMu.Lock()
	defer c.submittedJobsIdsMu.Unlock()
	c.submittedJobIds[id] = struct{}{}
	// Successfully submitted result
	// TODO: debug logger
	c.lastSubmittedShare = s
	return nil
}
