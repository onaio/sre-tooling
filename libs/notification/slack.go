package notification

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const slackWebhookURLEnvVar string = "SRE_NOTIFY_SLACK_WEBHOOK_URL"

type Slack struct {
	authToken string
}

func (snc Slack) SendMessage(message string) error {
	webhookURL := os.Getenv(slackWebhookURLEnvVar)
	if len(webhookURL) == 0 {
		return errors.New(fmt.Sprintf("%s not set", slackWebhookURLEnvVar))
	}
	var body = []byte(fmt.Sprintf(`{"text":"%s"}`, message))
	req, reqErr := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
	if reqErr != nil {
		return reqErr
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}

	return nil
}
