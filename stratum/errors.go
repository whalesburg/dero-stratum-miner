package stratum

import (
	"errors"
	"fmt"
)

type ErrorCode int

const (
	ERROR_UNKNOWN                ErrorCode = -1
	ERROR_SERVICE                ErrorCode = -2
	ERROR_METHOD                 ErrorCode = -3
	ERROR_FEE_REQUIRED           ErrorCode = -10
	ERROR_SIGNATURE_REQUIRED     ErrorCode = -20
	ERROR_SIGNATURE_UNAVAILABLE  ErrorCode = -21
	ERROR_UNKNOWN_SIGNATURE_TYPE ErrorCode = -22
	ERROR_BAD_SIGNATURE          ErrorCode = -23
)

type Error struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Traceback any       `json:"traceback"`
}

func (se *Error) Error() string {
	return fmt.Sprintf("code=%v msg=%v traceback=%v", se.Code, se.Message, se.Traceback)
}

var (
	ErrNoSessionID = errors.New("response has no session id")
	ErrNoJob       = errors.New("reponse has no job")
)
