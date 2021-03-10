package logutil

import (
	"io"
	"log"
	"strings"
)

// ToWriter converts a *log.Logger to an io.Writer
func ToWriter(logger *log.Logger) io.Writer {
	return &writerLogger{
		logger: logger,
	}
}

type writerLogger struct {
	logger *log.Logger
}

func (l writerLogger) Write(data []byte) (int, error) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line != "" {
			l.logger.Println(line)
		}
	}
	return len(data), nil
}
