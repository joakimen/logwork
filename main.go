package main

// Logs work on a Jira-task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	flag "github.com/spf13/pflag"
)

const API_TOKEN = "JIRA_API_TOKEN"
const JIRA_HOST= "JIRA_HOST"

type WorkLog struct {
	Comment          string `json:"comment"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}

func invalidArgExit(reason string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", reason)
	flag.Usage()
	os.Exit(1)
}

func main() {

	var (
		issue         = flag.StringP("issue-name", "i", "", "jira-issue (e.g. BCG-221)")
		comment       = flag.StringP("comment", "c", "", "worklog comment")
		minutesWorked = flag.IntP("minutes-worked", "t", 0, "time worked in minutes")
	)

	flag.Parse()

	if *issue == "" {
		invalidArgExit("issue-name cannot be blank.")
	}

	if *minutesWorked <= 0 || *minutesWorked >= (60*24) {
		invalidArgExit("minutes must be more than 0 and less than a full day.")
	}

	if *comment == "" {
		invalidArgExit("refusing to log work with empty message.")
	}

	apiToken := getEnvStrict(API_TOKEN)
	jiraHost := getEnvStrict(JIRA_HOST)

	worklog := WorkLog{
		Comment:          *comment,
		TimeSpentSeconds: *minutesWorked * 60,
	}

	api := fmt.Sprintf(jiraHost, *issue)

	jsonBody := jsonMarshal(worklog)
	request := newAuthenticatedRequest(api, jsonBody, apiToken)

	fmt.Printf("adding worklog for Jira issue: %s..\n", *issue)
	response := clientDo(request)
	body := parseBody(response)

	switch response.StatusCode {
	case http.StatusCreated:
		fmt.Printf("success: worklog was added for Jira issue: %s\n", *issue)
	case http.StatusBadRequest:
		log.Fatalf("invalid input: %v", body)
	case http.StatusForbidden:
		log.Fatalf("server returned: Unauthorized: %v", body)
	default:
		log.Fatalf("unhandled response: %d\n%s", response.StatusCode, body)
	}
}

func parseBody(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("couldn't parse response body: %v", err)
	}
	return string(body)
}

func clientDo(request *http.Request) *http.Response {
	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return response
}

func newAuthenticatedRequest(api string, jsonBody []byte, apiToken string) *http.Request {
	request, err := http.NewRequest(http.MethodPost, api, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiToken)
	return request
}

func jsonMarshal(worklog WorkLog) []byte {
	jsonBody, err := json.Marshal(worklog)
	if err != nil {
		log.Fatal(err)
	}
	return jsonBody
}

func getEnvStrict(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("env var is not set: %v", name)
	}
	return val
}
