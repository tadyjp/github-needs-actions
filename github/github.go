package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

type Config struct {
	AccessToken string
	Owner       string
	Repo        string
	Label       string
	DaysAgo     int
}

type RequestedPull struct {
	Pull  *github.PullRequest
	Users []*github.User
}

type RequestedPulls []RequestedPull

type RequestedIssue struct {
	Issue *github.Issue
	Users []*github.User
}

type RequestedIssues []RequestedIssue

func GetClient(cfg *Config) (*github.Client, *context.Context) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), &ctx
}

func GetRequestedPulls(client *github.Client, ctx *context.Context, cfg *Config) RequestedPulls {
	opt := &github.PullRequestListOptions{Direction: "asc"}

	var requestedPulls RequestedPulls

	for {
		pulls, resp, err := client.PullRequests.List(*ctx, cfg.Owner, cfg.Repo, opt)
		if err != nil {
			fmt.Println("GetPulls", err)
			os.Exit(1)
		}

		for _, pull := range pulls {
			if cfg.Label != "" {
				hasMatchLabel := false

				for _, label := range pull.Labels {
					if label.GetName() == cfg.Label {
						hasMatchLabel = true
					}
				}

				if !hasMatchLabel {
					continue
				}
			}
			reviewRequests, _, err := client.PullRequests.ListReviewers(*ctx, cfg.Owner, cfg.Repo, pull.GetNumber(), nil)
			if err != nil {
				fmt.Println("getReviews", err)
				os.Exit(1)
			}

			if len(reviewRequests.Users) == 0 {
				requestedPulls = append(requestedPulls, RequestedPull{Pull: pull, Users: []*github.User{pull.GetUser()}})
			} else {
				requestedPulls = append(requestedPulls, RequestedPull{Pull: pull, Users: reviewRequests.Users})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return requestedPulls
}

func (pulls *RequestedPulls) GetSlackText(cfg *Config) string {
	var sb strings.Builder

	if len(*pulls) == 0 {
		sb.WriteString("There is no waiting pull requests :tada:")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("These pull request (with %s tag) needs actions.", cfg.Label))
	sb.WriteString("\n")

	for _, pull := range *pulls {
		sb.WriteString("----------------------------------------")
		sb.WriteString("\n")

		sb.WriteString(fmt.Sprintf(":%s: <%s|%s#%d>: *%s* (<!date^%d^{date_short_pretty}|%s>)",
			strings.ToLower(pull.Pull.GetUser().GetLogin()),
			pull.Pull.GetHTMLURL(),
			pull.Pull.Base.GetRepo().GetFullName(),
			pull.Pull.GetNumber(),
			pull.Pull.GetTitle(),
			pull.Pull.GetCreatedAt().Unix(),
			pull.Pull.GetCreatedAt().String(),
		))
		sb.WriteString("\n")
		sb.WriteString("    Next Action: ")
		for _, user := range pull.Users {
			sb.WriteString(fmt.Sprintf(":%s: @%s", strings.ToLower(user.GetLogin()), user.GetLogin()))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func GetOldIssues(client *github.Client, ctx *context.Context, cfg *Config) RequestedIssues {
	opt := &github.IssueListByRepoOptions{Direction: "asc"}

	var requestedIssues RequestedIssues
	now := time.Now()

	for {
		issues, resp, err := client.Issues.ListByRepo(*ctx, cfg.Owner, cfg.Repo, opt)
		if err != nil {
			fmt.Println("ListByRepo", err)
			os.Exit(1)
		}

		for _, issue := range issues {
			if issue.GetUpdatedAt().After(now.Add(-time.Duration(cfg.DaysAgo*24) * time.Hour)) {
				continue
			}

			users := []*github.User{issue.GetUser()}

			if len(issue.Assignees) == 0 {
				users = append(users, issue.Assignees...)
			}

			requestedIssues = append(requestedIssues, RequestedIssue{Issue: issue, Users: users})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return requestedIssues
}

func (issues *RequestedIssues) GetSlackText(cfg *Config) string {
	var sb strings.Builder

	if len(*issues) == 0 {
		sb.WriteString("There is no old issues :tada:")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("These issues (no update within %d days) needs actions.", cfg.DaysAgo))
	sb.WriteString("\n")

	for _, issue := range *issues {
		sb.WriteString("----------------------------------------")
		sb.WriteString("\n")

		sb.WriteString(fmt.Sprintf(":%s: <%s|%s#%d>: *%s* (<!date^%d^{date_short_pretty}|%s> -> <!date^%d^{date_short_pretty}|%s>)",
			strings.ToLower(issue.Issue.GetUser().GetLogin()),
			issue.Issue.GetHTMLURL(),
			issue.Issue.GetRepository().GetFullName(),
			issue.Issue.GetNumber(),
			issue.Issue.GetTitle(),
			issue.Issue.GetCreatedAt().Unix(),
			issue.Issue.GetCreatedAt().String(),
			issue.Issue.GetUpdatedAt().Unix(),
			issue.Issue.GetUpdatedAt().String(),
		))
		sb.WriteString("\n")
		sb.WriteString("    Next Action: ")
		for _, user := range issue.Users {
			sb.WriteString(fmt.Sprintf(":%s: @%s", strings.ToLower(user.GetLogin()), user.GetLogin()))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
