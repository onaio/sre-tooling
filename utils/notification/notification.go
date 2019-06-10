package notification

type Channel interface {
	SendMessage(message string) error
}

func GetActiveChannels() []Channel {
	channels := []Channel{}
	channels = append(channels, Log{})
	channels = append(channels, Slack{})

	return channels
}

func SendMessage(message string) error {
	channels := GetActiveChannels()
	for _, curChannel := range channels {
		curErr := curChannel.SendMessage(message)
		if curErr != nil {
			return curErr
		}
	}

	return nil
}
