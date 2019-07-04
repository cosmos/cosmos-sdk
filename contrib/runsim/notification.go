package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nlopes/slack"
)

const (
	logBucketPrefix = "sim-logs-"
	awsRegion       = "us-east-1"
)

func awsErrHandler(err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		default:
			return awsErr
		}
	}
	return err
}

func makeObjKey(objKeyPrefix string, fileName string) string {
	return fmt.Sprintf("%s/%s", objKeyPrefix, fileName)
}

// putObjects attempts to upload to an S3 bucket the content of each file from fileHandles.
// File descriptors have their read offset set to 0 to ensure all the content is uploaded.
// Each file will become an S3 bucket object that can be accessed via its object key.
//
// Function returns the list of object keys and an error, if any.
func putObjects(svc *s3.S3, objKeyPrefix string, bucketName string, fileHandles ...*os.File) ([]string, error) {
	objKeys := make([]string, len(fileHandles))
	for index, fileHandle := range fileHandles {
		_, _ = fileHandle.Seek(0, 0)
		objKey := makeObjKey(objKeyPrefix, filepath.Base(fileHandle.Name()))
		stdOutObjInput := &s3.PutObjectInput{
			Body:   aws.ReadSeekCloser(fileHandle),
			Bucket: aws.String(bucketName),
			Key:    aws.String(objKey),
		}
		_, err := svc.PutObject(stdOutObjInput)
		if err != nil {
			return nil, awsErrHandler(err)
		}
		objKeys[index] = objKey
	}
	return objKeys, nil
}

func pushLogs(stdOut *os.File, stdErr *os.File, folderName string) ([]string, *string, error) {
	var logBucket *string

	sessionS3 := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})))
	listBucketsOutput, err := sessionS3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, nil, awsErrHandler(err)
	}
	for _, bucket := range listBucketsOutput.Buckets {
		if strings.Contains(*bucket.Name, logBucketPrefix) {
			logBucket = bucket.Name
			objKeys, err := putObjects(sessionS3, folderName, *logBucket, stdOut, stdErr)
			if err != nil {
				return nil, nil, err
			}
			return objKeys, bucket.Name, nil
		}
	}
	return nil, nil, nil
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

//type GithubPayload struct {
//	Issue struct {
//		Number int `json:"number"`
//		Pull   struct {
//			Url string `json:"url,omitempty"`
//		} `json:"pull_request,omitempty"`
//	} `json:"issue"`
//
//	Comment struct {
//		Body string `json:"body"`
//	} `json:"comment"`
//
//	Repository struct {
//		Name  string `json:"name"`
//		Owner struct {
//			Login string `json:"login"`
//		} `json:"owner"`
//	} `json:"repository"`
//}
//
//type PullRequestDetails struct {
//	Head struct {
//		Ref string `json:"ref"`
//		Sha string `json:"sha"`
//	} `json:"head"`
//}

//func createCheckRun(client *github.Client, payload GithubPayload, pr PullRequestDetails) error {
//	var opt github.CreateCheckRunOptions
//	opt.Name = "Test Check"
//	opt.HeadBranch = pr.Head.Ref
//	opt.HeadSHA = pr.Head.Sha
//
//	checkRUn, resp, err := client.Checks.CreateCheckRun(context.Background(), payload.Repository.Owner.Login, payload.Repository.Name, opt)
//	log.Printf("%v", resp)
//	log.Printf("%v", checkRUn)
//	if err != nil {
//		log.Printf("ERROR: CreateCheckRun: %v", err.Error())
//		return err
//	}
//	return err
//}
//
//func getPrDetails(prUrl string) (*PullRequestDetails, error) {
//	request, err := http.Get(prUrl)
//	if err != nil {
//		return nil, err
//	}
//
//	var details PullRequestDetails
//	if err := json.NewDecoder(request.Body).Decode(&details); err != nil {
//		return nil, err
//	}
//
//	return &details, nil
//}
//
//func updateCheckRun(client *github.Client, payload GithubPayload, pr PullRequestDetails) error {
//	status := "completed"
//	conclusion := "success"
//	var opt github.UpdateCheckRunOptions
//	opt.Name = "Test Check"
//	opt.Status = &status
//	opt.Conclusion = &conclusion
//	ts := github.Timestamp{Time: time.Now()}
//	opt.CompletedAt = &ts
//
//	updatedCheck, resp, err := client.Checks.UpdateCheckRun(context.Background(), payload.Repository.Owner.Login, payload.Repository.Name, 136693316, opt)
//	log.Printf("%v", updatedCheck)
//	log.Printf("%v", resp)
//	if err != nil {
//		log.Printf("ERROR: UpdateCheckRun: %v", err.Error())
//		return err
//	}
//	return nil
//}

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
