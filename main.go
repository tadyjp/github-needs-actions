package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"

	"github.com/tadyjp/github-needs-actions/github"
	"github.com/tadyjp/github-needs-actions/slack"
)

func run() {
	githubConfig := &github.Config{
		AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
		Owner:       os.Getenv("GITHUB_TARGET_OWNER"),
		Repo:        os.Getenv("GITHUB_TARGET_REPO"),
		Label:       os.Getenv("GITHUB_TARGET_LABEL"),
	}

	if d, ok := os.LookupEnv("GITHUB_DAYS_AGO"); ok {
		if i, err := strconv.Atoi(d); err == nil {
			githubConfig.DaysAgo = i
		}
	} else {
		githubConfig.DaysAgo = 0
	}

	slackConfig := &slack.Config{
		WebhookURL: os.Getenv("SLACK_WEBHOOKURL"),
		Channel:    os.Getenv("SLACK_CHANNEL"),
		Username:   os.Getenv("SLACK_USERNAME"),
	}

	client, ctx := github.GetClient(githubConfig)

	requestedPulls := github.GetRequestedPulls(client, ctx, githubConfig)
	pullsSlackText := requestedPulls.GetSlackText(githubConfig)
	fmt.Println(pullsSlackText)
	slack.PostTextToSlack(slackConfig, pullsSlackText)

	requestedIssues := github.GetOldIssues(client, ctx, githubConfig)
	issuesSlackText := requestedIssues.GetSlackText(githubConfig)
	fmt.Println(issuesSlackText)
	slack.PostTextToSlack(slackConfig, issuesSlackText)
}

func main() {
	if os.Getenv("DEBUG") == "1" {
		fmt.Println("Running in debug mode.")
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		run()
	} else {
		lambda.Start(run)
	}
}
