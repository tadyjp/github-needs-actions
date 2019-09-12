package main

import (
	"os"

	"github.com/ashwanthkumar/slack-go-webhook"
)

var (
	SLACK_WEBHOOKURL = os.Getenv("SLACK_WEBHOOKURL")
	SLACK_CHANNEL = os.Getenv("SLACK_CHANNEL")
	SLACK_USERNAME = os.Getenv("SLACK_USERNAME")
)

func PostSlack(msg string) {
	field1 := slack.Field{Value: msg}

	attachment := slack.Attachment{}
	attachment.AddField(field1)
	color := "good"
	attachment.Color = &color
	payload := slack.Payload{
		Username:    SLACK_USERNAME,
		Channel:     SLACK_CHANNEL,
		Attachments: []slack.Attachment{attachment},
		Markdown:    true,
	}
	err := slack.Send(SLACK_WEBHOOKURL, "", payload)
	if err != nil {
		os.Exit(1)
	}
}
