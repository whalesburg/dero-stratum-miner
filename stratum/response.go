package stratum

type Response struct {
	ID     any            `json:"id"`
	Result map[string]any `json:"result"`
	Error  *Error         `json:"error"`
}
