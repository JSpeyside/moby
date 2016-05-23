package infrastructure

import (
	"fmt"
	// "time"
)

type Logger struct {
}

func (logger Logger) Log(message string) error {
	// now := time.Now()
	// fmt.Printf("%s %s\n", now.Format(time.RFC3339), message)
	fmt.Println(message)
	return nil
}

func (logger Logger) LogLines(messages []string) error {
	for _, message := range messages {
		logger.Log(message)
	}
	return nil
}

func NewLogger() *Logger {
	return &Logger{}
}
