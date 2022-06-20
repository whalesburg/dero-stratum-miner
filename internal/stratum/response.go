package stratum

type Response struct {
	ID     any    `json:"id"`
	Result any    `json:"result"`
	Error  *Error `json:"error"`
}
