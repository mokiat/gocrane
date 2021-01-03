package project

import (
	"log"
	"strings"
)

type logWriter struct {
	logger *log.Logger
}

func (w logWriter) Write(data []byte) (int, error) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line != "" {
			w.logger.Println(line)
		}
	}
	return len(data), nil
}
