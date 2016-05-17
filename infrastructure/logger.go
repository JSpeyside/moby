package infrastructure

import (
	"fmt"
	"time"
)

type Logger struct {
}

func (logger Logger) Log(message string) error {
	now := time.Now()
	fmt.Printf("%s %s", now.Format(time.RFC3339), message)
	return nil
}
