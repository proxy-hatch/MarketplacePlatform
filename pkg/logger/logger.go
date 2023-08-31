package logger

import (
    "errors"
    "fmt"
    "go.uber.org/zap"
    "log"
    "syscall"
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
        err := logger.Sync()
        if err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EINVAL) {
            fmt.Println(err)
        }
    }(logger) // flushes buffer, if any
    return logger.Sugar()
}
