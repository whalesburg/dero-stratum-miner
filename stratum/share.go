package stratum

import (
	"context"
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

func (s *Share) Encode() (map[string]any, error) {
	return map[string]any{
		"id":     s.ID,
		"job_id": s.JobID,
		"nonce":  s.Nonce,
		"result": s.Result,
	}, nil
}

func (c *Client) SubmitShare(ctx context.Context, s *Share) error {
	if s.JobID == c.lastSubmittedShare.JobID {
		c.LogFn.Error(fmt.Errorf("duplicate share %s", s.JobID), fmt.Sprintf("job id %s", s.JobID))
		return nil
	}

	var res any
	params, err := s.Encode()
	if err != nil {
		return err
	}

	err = c.client.CallWithContext(ctx, "submit", params, &res)
	if err != nil {
		return err
	}
	fmt.Println(res)
	/* id, ok := req.ID.(int)
	if !ok {
		return fmt.Errorf("failed to convert id to int: %v", req.ID)
	} */
	c.submittedJobsIdsMu.Lock()
	defer c.submittedJobsIdsMu.Unlock()
	//c.submittedJobIds[id] = struct{}{}
	// Successfully submitted result
	// TODO: debug logger
	c.lastSubmittedShare = s
	return nil
}
