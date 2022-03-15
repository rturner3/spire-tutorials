package logger

import (
	"io"

	zaphook "github.com/Sytten/logrus-zap-hook"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func NewLogrusBridgeLogger(logger *zap.Logger) logrus.FieldLogger {
	logrusLogger := logrus.New()
	logrusLogger.ReportCaller = true
	logrusLogger.SetOutput(io.Discard)

	hook, _ := zaphook.NewZapHook(logger)
	logrusLogger.Hooks.Add(hook)
	return logrusLogger
}
