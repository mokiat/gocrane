package project

import "log"

type logWriter struct {
	logger *log.Logger
}

func (w logWriter) Write(data []byte) (int, error) {
	w.logger.Print(string(data))
	return len(data), nil
}
