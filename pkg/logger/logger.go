package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "log"
    "os"
    "path/filepath"
)

func NewLogger() *zap.SugaredLogger {
    // Create the log directory if it does not exist
    err := os.MkdirAll("log", 0755)
    if err != nil {
        log.Fatalf("can't create log directory: %v", err)
    }

    // Open the log file in append mode, or create it if it does not exist
    file, err := os.OpenFile(filepath.Join("log", "application.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("can't open log file: %v", err)
    }

    // Create a zapcore.Core that writes to the file
    core := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
        // TODO: use JSON encoder in production
        // zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(file),
        zapcore.DebugLevel,
    )

    // Create a zap.Logger with the core
    logger := zap.New(core)

    // defer func(logger *zap.Logger) {
    //     logger.Sync()
    // }(logger) // flushes buffer, if any
    //
    return logger.Sugar()
}
