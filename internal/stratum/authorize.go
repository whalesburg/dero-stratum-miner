package stratum

func (c *Client) authorize() error {
	args := map[string]any{
		"login": c.username,
		"pass":  c.password,
		"agent": c.agentName,
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

	sid, ok := response.Result.(map[string]any)
	if !ok {
		return ErrNoSessionID
	}
	c.sessionID, ok = sid["id"].(string)
	if !ok {
		return ErrNoSessionID
	}
	job, err := extractJob(response.Result.(map[string]any)["job"].(map[string]any))
	if err != nil {
		return err
	}
	c.broadcastJob(job)

	return nil
}

func (c *Client) readResponse() (*Response, error) {
	line, err := c.readLine()
	if err != nil {
		return nil, err
	}
	return parseResponse(line)
}
