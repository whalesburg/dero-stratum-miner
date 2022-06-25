package stratum

import (
	"errors"
	"fmt"
)

type ErrorCode int

const (
	ErrUnknown              ErrorCode = -1
	ErrService              ErrorCode = -2
	ErrMethod               ErrorCode = -3
	ErrFeeRequired          ErrorCode = -10
	ErrSignatureRequired    ErrorCode = -20
	ErrSignatureUnavailable ErrorCode = -21
	ErrUnkownSignatureType  ErrorCode = -22
	ErrBadSignature         ErrorCode = -23
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
