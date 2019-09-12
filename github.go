package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v28/github"
)

var (
	GITHUB_ACCESS_TOKEN = os.Getenv("GITHUB_ACCESS_TOKEN")
	GITHUB_TARGET_OWNER = os.Getenv("GITHUB_TARGET_OWNER")
	GITHUB_TARGET_REPO = os.Getenv("GITHUB_TARGET_REPO")
	GITHUB_TARGET_LABEL = os.Getenv("GITHUB_TARGET_LABEL")
)

type RequestedPull struct {
	Pull  *github.PullRequest
	Users []*github.User
}

func getRequestedPulls(client *github.Client, ctx *context.Context) []*RequestedPull {
	opt := &github.PullRequestListOptions{Direction: "asc"}
	pulls, _, err := client.PullRequests.List(*ctx, GITHUB_TARGET_OWNER, GITHUB_TARGET_REPO, opt)
	if err != nil {
		fmt.Println("GetPulls", err)
		os.Exit(1)
	}

	var requestedPulls []*RequestedPull

	for _, pull := range pulls {
		if len(GITHUB_TARGET_LABEL) > 0 {
			hasMatchLabel := false

			for _, label := range pull.Labels {
				if label.GetName() == GITHUB_TARGET_LABEL {
					hasMatchLabel = true
				}
			}

			if !hasMatchLabel {
				continue
			}
		}
		reviewRequests, _, err := client.PullRequests.ListReviewers(*ctx, GITHUB_TARGET_OWNER, GITHUB_TARGET_REPO, pull.GetNumber(), nil)
		if err != nil {
			fmt.Println("getReviews", err)
			os.Exit(1)
		}

		if len(reviewRequests.Users) == 0 {
			requestedPulls = append(requestedPulls, &RequestedPull{Pull: pull, Users: []*github.User{pull.GetUser()}})
		} else {
			requestedPulls = append(requestedPulls, &RequestedPull{Pull: pull, Users: reviewRequests.Users})
		}
	}

	return requestedPulls
}

func getClient() (*github.Client, *context.Context) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), &ctx
}

func getSlackText(requestedPulls []*RequestedPull) string {
	var sb strings.Builder

	if len(requestedPulls) == 0 {
		sb.WriteString("There is no waiting pull requests :tada:")
		return sb.String()
	}

	for i, requestedPull := range requestedPulls {
		if i != 0 {
			sb.WriteString("----------------------------------------")
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf(":%s: <%s|%s#%d>: *%s* (<!date^%d^{date_short_pretty}|%s>)",
			strings.ToLower(requestedPull.Pull.GetUser().GetLogin()),
			requestedPull.Pull.GetHTMLURL(),
			requestedPull.Pull.Base.GetRepo().GetFullName(),
			requestedPull.Pull.GetNumber(),
			requestedPull.Pull.GetTitle(),
			requestedPull.Pull.GetCreatedAt().Unix(),
			requestedPull.Pull.GetCreatedAt().String(),
		))
		sb.WriteString("\n")
		sb.WriteString("    Next Action: ")
		for _, user := range requestedPull.Users {
			sb.WriteString(fmt.Sprintf(":%s: @%s", strings.ToLower(user.GetLogin()), user.GetLogin()))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func PostToSlack() {
	client, ctx := getClient()

	requestedPulls := getRequestedPulls(client, ctx)

	slackText := getSlackText(requestedPulls)

	fmt.Println(slackText)

	PostSlack(slackText)
}
