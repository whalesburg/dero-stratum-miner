package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
)

type Job struct {
	ID         string  `json:"job_id"`
	Blob       string  `json:"blob"`
	Height     float64 `json:"height"`
	ExtraNonce string  `json:"extra_nonce"`
	PoolWallet string  `json:"pool_wallet"`
	Target     string  `json:"target"`
	Difficulty uint64
}

func extractJob(data map[string]any) (*Job, error) {
	if data == nil {
		return nil, ErrNoJob
	}

	var (
		job Job
		ok  bool
	)
	job.ID, ok = data["job_id"].(string)
	if !ok {
		return nil, errors.New("failed to cast job_id")
	}
	job.Blob, ok = data["blob"].(string)
	if !ok {
		return nil, errors.New("failed to cast blob")
	}
	job.Height, ok = data["height"].(float64)
	if !ok {
		return nil, errors.New("failed to cast height")
	}
	job.ExtraNonce, ok = data["extra_nonce"].(string)
	if !ok {
		return nil, errors.New("failed to cast extra nonce")
	}
	job.PoolWallet, ok = data["pool_wallet"].(string)
	if !ok {
		return nil, errors.New("failed to cast pool wallet")
	}
	job.Target, ok = data["target"].(string)
	if !ok {
		return nil, errors.New("failed to cast target")
	}

	raw, err := hex.DecodeString(job.Target)
	if err != nil {
		return nil, errors.New("failed to decode target")
	}
	var a = binary.LittleEndian.Uint64(raw)
	job.Difficulty = uint64(0xFFFFFFFFFFFFFFFF / a)

	return &job, nil

}

func (c *Client) broadcastJob(job *Job) {
	c.jobBroadcaster.Notify(job)
}
