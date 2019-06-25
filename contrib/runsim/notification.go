package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v26/github"
	"github.com/nlopes/slack"
)

type GithubPayload struct {
	Issue struct {
		Number int `json:"number"`
		Pull   struct {
			Url string `json:"url,omitempty"`
		} `json:"pull_request,omitempty"`
	} `json:"issue"`

	Comment struct {
		Body string `json:"body"`
	} `json:"comment"`

	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

type PullRequestDetails struct {
	Head struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	} `json:"head"`
}

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

func createCheckRun(client *github.Client, payload GithubPayload, pr PullRequestDetails) error {
	var opt github.CreateCheckRunOptions
	opt.Name = "Test Check"
	opt.HeadBranch = pr.Head.Ref
	opt.HeadSHA = pr.Head.Sha

	checkRUn, resp, err := client.Checks.CreateCheckRun(context.Background(), payload.Repository.Owner.Login, payload.Repository.Name, opt)
	log.Printf("%v", resp)
	log.Printf("%v", checkRUn)
	if err != nil {
		log.Printf("ERROR: CreateCheckRun: %v", err.Error())
		return err
	}
	return err
}

func getPrDetails(prUrl string) (*PullRequestDetails, error) {
	request, err := http.Get(prUrl)
	if err != nil {
		return nil, err
	}

	var details PullRequestDetails
	if err := json.NewDecoder(request.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

func updateCheckRun(client *github.Client, payload GithubPayload, pr PullRequestDetails) error {
	status := "completed"
	conclusion := "success"
	var opt github.UpdateCheckRunOptions
	opt.Name = "Test Check"
	opt.Status = &status
	opt.Conclusion = &conclusion
	ts := github.Timestamp{Time: time.Now()}
	opt.CompletedAt = &ts

	updatedCheck, resp, err := client.Checks.UpdateCheckRun(context.Background(), payload.Repository.Owner.Login, payload.Repository.Name, 136693316, opt)
	log.Printf("%v", updatedCheck)
	log.Printf("%v", resp)
	if err != nil {
		log.Printf("ERROR: UpdateCheckRun: %v", err.Error())
		return err
	}
	return nil
}

//func githubCheckHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
//	response := events.APIGatewayProxyResponse{StatusCode: 200}
//	var comment GithubPayload
//	if err := json.NewDecoder(bytes.NewBufferString(request.Body)).Decode(&comment); err != nil {
//		response.StatusCode = 500
//		response.Body = err.Error()
//		return response, err
//	}
//
//	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 30867, 997580, "github-integration/gaia-sim.2019-05-16.private-key.pem")
//	if err != nil {
//		response.StatusCode = 500
//		response.Body = err.Error()
//		log.Printf("AuthError: %v", err)
//		return response, err
//	}
//	client := github.NewClient(&http.Client{Transport: itr})
//	message := "App comment"
//	issue := new(github.IssueComment)
//	issue.Body = &message
//
//	if comment.Comment.Body == "Start sim" && comment.Issue.Pull.Url != "" {
//		prDetails, err := getPrDetails(comment.Issue.Pull.Url)
//		if err != nil {
//			response.StatusCode = 500
//			response.Body = err.Error()
//			log.Printf("ERROR: getPrDetails: %v", err.Error())
//			return response, err
//		}
//		log.Printf("%v", prDetails)
//
//		if err := createCheckRun(client, comment, *prDetails); err != nil {
//			response.StatusCode = 500
//			response.Body = err.Error()
//			return response, err
//		}
//
//		comments, resp, err := client.Issues.CreateComment(context.Background(),
//			comment.Repository.Owner.Login, comment.Repository.Name, comment.Issue.Number, issue)
//
//		log.Printf("%v", resp)
//		log.Printf("%v", comments)
//		if err != nil {
//			log.Printf("ERROR: CreateComment: %v", err.Error())
//			response.StatusCode = 500
//			response.Body = err.Error()
//			return response, err
//		}
//	}
//
//	if comment.Comment.Body == "Update check" && comment.Issue.Pull.Url != "" {
//		prDetails, err := getPrDetails(comment.Issue.Pull.Url)
//		if err != nil {
//			response.StatusCode = 500
//			response.Body = err.Error()
//			log.Printf("ERROR: getPrDetails: %v", err.Error())
//			return response, err
//		}
//		log.Printf("%v", prDetails)
//
//		if err := updateCheckRun(client, comment, *prDetails); err != nil {
//			response.StatusCode = 500
//			response.Body = err.Error()
//			log.Printf("ERROR: getPrDetails: %v", err.Error())
//			return response, err
//		}
//	}
//
//	return response, nil
//}
