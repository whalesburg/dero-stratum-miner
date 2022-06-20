package stratum

import (
	"encoding/json"
)

type Request struct {
	ID     any    `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

func NewRequest(id int, method string, args any) *Request {
	return &Request{
		id,
		method,
		args,
	}
}

func (r *Request) Parse() ([]byte, error) {
	payload := make(map[string]any)
	payload["jsonrpc"] = "2.0"
	payload["method"] = r.Method
	payload["id"] = r.ID
	payload["params"] = r.Params

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 0, len(data)+1)
	b = append(b, data...)
	b = append(b, '\n')
	return b, nil
}
