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
	text := strings.TrimSpace(string(data))
	for line := range strings.SplitSeq(text, "\n") {
		l.logger.Println(line)
	}
	return len(data), nil
}
