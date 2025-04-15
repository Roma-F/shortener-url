package logger

import "go.uber.org/zap"

var Sugar *zap.SugaredLogger

func Initialize(level string) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	Sugar = logger.Sugar()
}
