package notification

import (
	"fmt"
	"strings"
)

type Log struct {
}

func (snc Log) SendMessage(message string) error {
	splitString := strings.Split(message, "\n")
	for _, curLine := range splitString {
		if len(curLine) > 0 {
			fmt.Println(curLine)
		}
	}
	return nil
}
