package notify

import (
	"fmt"
	"strings"
)

type LogNotificationChannel struct {
}

func (snc LogNotificationChannel) SendMessage(message string) error {
	splitString := strings.Split(message, "\n")
	for _, curLine := range splitString {
		if len(curLine) > 0 {
			fmt.Println(curLine)
		}
	}
	return nil
}
