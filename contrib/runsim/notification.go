package main

import (
	"github.com/nlopes/slack"
	"log"
)

func slackMessage(token string, channel string, threadTS *string, message string) {
	client := slack.New(token)
	if threadTS != nil {
		_, _, err := client.PostMessage(channel, slack.MsgOptionText(message, false), slack.MsgOptionTS(*threadTS))
		if err != nil {
			log.Printf("ERROR: %v", err)
		}
	} else {
		_, _, err := client.PostMessage(channel, slack.MsgOptionText(message, false))
		if err != nil {
			log.Printf("ERROR: %v", err)
		}
	}

}
