package logger

import (
    "go.uber.org/zap"
    "log"
)

func NewLogger() *zap.SugaredLogger {
    // TODO: print to log file instead of stdout

    // TODO: disable DEBUG level in prod
    // logger, err := zap.NewProduction()
    logger, err := zap.NewDevelopment()
    if err != nil {
        log.Fatalf("can't initialize zap logger: %v", err)
    }

    defer func(logger *zap.Logger) {
        logger.Sync()
    }(logger) // flushes buffer, if any
    return logger.Sugar()
}
