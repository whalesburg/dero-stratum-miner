package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type MiningAuthorizeRequest struct {
	username  string
	password  string
	agentName string
}

func (s *MiningAuthorizeRequest) Encode() (map[string]any, error) {
	if s.username == "" {
		return nil, fmt.Errorf("Empty user in authorize request")
	}
	return map[string]any{
		"login": s.username,
		"pass":  s.password,
		"agent": s.agentName,
	}, nil
}

type Job struct {
	ID         string  `json:"job_id"`
	Blob       string  `json:"blob"`
	Height     float64 `json:"height"`
	ExtraNonce string  `json:"extra_nonce"`
	PoolWallet string  `json:"pool_wallet"`
	Target     string  `json:"target"`
	Difficulty uint64
}

var (
	ErrNoSessionID = errors.New("response has no session id")
	ErrNoJob       = errors.New("reponse has no job")
)

func (j *Job) Decode(data map[string]any) error {
	var ok bool
	j.ID, ok = data["job_id"].(string)
	if !ok {
		return errors.New("failed to cast job_id")
	}
	j.Blob, ok = data["blob"].(string)
	if !ok {
		return errors.New("failed to cast blob")
	}
	j.Height, ok = data["height"].(float64)
	if !ok {
		return errors.New("failed to cast height")
	}
	j.ExtraNonce, ok = data["extra_nonce"].(string)
	if !ok {
		return errors.New("failed to cast extra nonce")
	}
	j.PoolWallet, ok = data["pool_wallet"].(string)
	if !ok {
		return errors.New("failed to cast pool wallet")
	}
	j.Target, ok = data["target"].(string)
	if !ok {
		return errors.New("failed to cast target")
	}

	raw, err := hex.DecodeString(j.Target)
	if err != nil {
		return errors.New("failed to decode target")
	}
	var a = binary.LittleEndian.Uint64(raw)
	if a == 0 {
		return errors.New("invalid target")
	}
	j.Difficulty = 0xFFFFFFFFFFFFFFFF / a

	return nil
}
