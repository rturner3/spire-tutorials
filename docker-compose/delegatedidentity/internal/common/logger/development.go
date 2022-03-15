package logger

import "go.uber.org/zap"

func NewDevelopment() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
