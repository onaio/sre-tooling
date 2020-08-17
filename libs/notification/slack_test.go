package notification

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSendMessage(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Check HTTP Method
		if req.Method != "POST" {
			t.Errorf("Unexpected HTTP Method: %s. Expected POST", req.Method)
		}
		// Check request Body
		expectedBody := `{"text":"hello slack"}`
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		actualBody := buf.String()
		if actualBody != expectedBody {
			t.Errorf("Unexepcted Body: %s. Expected %s", req.Body, expectedBody)
		}
	}))
	os.Setenv("SRE_NOTIFICATION_SLACK_WEBHOOK_URL", testServer.URL)
	slackChannel := Slack{}
	slackChannel.SendMessage("hello slack")
}
