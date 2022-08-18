package logging

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/whalesburg/dero-stratum-miner/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(console, logfile io.Writer, cfg *config.Logger) logr.Logger {
	var logLevelConsole zap.AtomicLevel
	if cfg.Debug { // setup debug mode if requested
		cfg.CLogLevel = 1
	}

	if cfg.CLogLevel < 0 {
		cfg.CLogLevel = 0
	}
	if cfg.CLogLevel > 127 {
		cfg.CLogLevel = 127
	}
	logLevelConsole = zap.NewAtomicLevelAt(zapcore.Level(0 - cfg.CLogLevel))

	zc := zap.NewDevelopmentEncoderConfig()
	zc.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zc.EncodeTime = zapcore.TimeEncoderOfLayout("02/01 15:04:05")

	consoleEncoder := zapcore.NewConsoleEncoder(zc)

	coreConsole := zapcore.NewCore(consoleEncoder, zapcore.AddSync(console), logLevelConsole)
	removecore := &removeCallerCore{coreConsole}
	core := zapcore.NewTee(
		removecore,
	)

	zcore := zap.New(core, zap.AddCaller()) // add caller info to every record which is then trimmed from console
	return zapr.NewLogger(zcore)
}

// remove caller information from console
type removeCallerCore struct {
	zapcore.Core
}

func (c *removeCallerCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Core.Check(entry, nil) == nil {
		return ce
	}
	return ce.AddCore(entry, c)
}

func (c *removeCallerCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	entry.Caller = zapcore.EntryCaller{}
	return c.Core.Write(entry, fields)
}

func (c *removeCallerCore) With(fields []zap.Field) zapcore.Core {
	return &removeCallerCore{c.Core.With(fields)}
}
