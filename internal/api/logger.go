package api

import (
	"fmt"

	"github.com/go-logr/logr"
)

type logger struct {
	logr logr.Logger
}

func (l *logger) Logf(format string, args ...interface{}) {
	l.logr.V(1).Info(fmt.Sprintf(format, args...))
}
