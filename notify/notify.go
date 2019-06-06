package notify

type NotificationChannel interface {
	SendMessage(message string) error
}

func GetActiveNotificationChannels() []NotificationChannel {
	channels := []NotificationChannel{}
	channels = append(channels, LogNotificationChannel{})
	channels = append(channels, SlackNotificationChannel{})

	return channels
}

func SendMessage(message string) error {
	channels := GetActiveNotificationChannels()
	for _, curChannel := range channels {
		curErr := curChannel.SendMessage(message)
		if curErr != nil {
			return curErr
		}
	}

	return nil
}
