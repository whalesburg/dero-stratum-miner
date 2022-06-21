package stratum

type Report struct {
	Hashrate uint64
}

func NewReport(hashrate uint64) *Report {
	return &Report{
		Hashrate: hashrate,
	}
}

func (c *Client) ReportHashrate(r *Report) error {
	args := make(map[string]interface{})
	args["id"] = c.sessionID
	args["hashrate"] = r.Hashrate
	_, err := c.call("reported_hashrate", args)
	if err != nil {
		return err
	}
	return nil
}
