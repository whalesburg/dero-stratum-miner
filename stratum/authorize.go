package stratum

import (
	"encoding/json"
)

func (c *Client) Authorize() error {
	args := map[string]any{
		"login": c.username,
		"pass":  c.password,
		"agent": "go-stratum",
	}

	if _, err := c.call("login", args); err != nil {
		return err
	}

	response, err := c.readResponse()
	if err != nil {
		return err
	}
	if response.Error != nil {
		return response.Error
	}
	c.connected = true

	sid, ok := response.Result["id"]
	if !ok {
		return ErrNoSessionID
	}
	c.sessionID = sid.(string)
	job, err := extractJob(response.Result["job"].(map[string]any))
	if err != nil {
		return err
	}
	c.broadcastJob(job)

	// Handle messages
	go c.handleMessages()
	return nil
}

func (c *Client) readResponse() (*Response, error) {
	line, err := c.readLine()
	if err != nil {
		return nil, err
	}
	return parseResponse(line)
}

func (c *Client) readLine() ([]byte, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	//fmt.Println("Received:", string(line))
	return line, nil
}

func parseResponse(b []byte) (*Response, error) {
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
