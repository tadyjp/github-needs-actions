package slack

import (
	"fmt"
	"os"

	"github.com/ashwanthkumar/slack-go-webhook"
)

type Config struct {
	WebhookURL string
	Channel    string
	Username   string
}

func PostTextToSlack(cfg *Config, msg string) {
	payload := slack.Payload{
		Username: cfg.Username,
		Channel:  cfg.Channel,
		Markdown: true,
		
		Text:     msg,
	}

	err := slack.Send(cfg.WebhookURL, "", payload)
	if err != nil {
		fmt.Println("PostTextToSlack", err)
		os.Exit(1)
	}
}
