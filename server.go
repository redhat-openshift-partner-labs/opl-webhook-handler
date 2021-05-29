package main

// Authenticate with Github via go-github library
// Get LabRequest (latest open issue with triage/accepted label from opdev/lab-requests)
// Validate LabRequest
// Create LabRequestBranch for LabRequest via CreateBranch
// Create LabRequestFile

import (
	"flag"
	. "github.com/rhecoeng/utils"
	whgh "gopkg.in/go-playground/webhooks.v5/github"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	path = "/"
)

func main() {
	org := flag.String("org", "rhecoeng", "Github Organization")
	repo := flag.String("repo", "opl-requests", "Github Repository")
	branch := flag.String("branch", "master", "Github Branch")
	whsecret := flag.String("whsecret", "", "Github Webhook Secret")

	flag.Parse()

	whenvsecret, ok := os.LookupEnv("GITHUB_SECRET")
	if ok {
		*whsecret = whenvsecret
	}

	hook, _ := whgh.New(whgh.Options.Secret(*whsecret))
	client, ctx := GithubAuthenticate()
	clientset := K8sAuthenticate()

	var labrequest LabRequest

	// Get commit SHA for master and store it in CurrentLabRequestBranch.Base field
	var CurrentLabRequestBranch LabRequestBranch
	masterBranch, _, err := client.Repositories.GetBranch(ctx, *org, *repo, *branch)
	ErrorCheck("Unable to get the "+*branch+" branch for SHA", err)
	CurrentLabRequestBranch.Base = *masterBranch.GetCommit().SHA

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, whgh.PullRequestEvent, whgh.IssuesEvent)

		if err != nil {
			if err == whgh.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
				log.Printf("%v is not an event being watched.\n", r.Header.Get("X-GitHub-Event"))
			}
		}

		switch payload.(type) {
		case whgh.PullRequestPayload:
			pr := payload.(whgh.PullRequestPayload)
			// Delete branch of PR that is closed and merged
			if pr.Action == "closed" && pr.PullRequest.Merged == true {
				DeleteBranch(&CurrentLabRequestBranch, client, ctx)
				CreateClusterDeployment(&labrequest)
			}

		case whgh.IssuesPayload:
			issue := payload.(whgh.IssuesPayload)
			// Create PR for triage/accepted lab request Issue
			if issue.Action == "labeled" && issue.Label.Name == "triage/accepted" {
				// Validate the labrequest
				// Validate also creates the json file for PR upon successful validation
				labrequest = Validate(issue.Issue.Body)

				// After validation create OpenShift secret to store information for lab
				CreateLabSecret(clientset, &labrequest)

				// Assign labrequest.ID to CurrentLabRequestBranch.Lab
				CurrentLabRequestBranch.Lab = labrequest.ID.String()

				// Generate default SSH key for lab-request and store in LabRequest.PublicSSHKey
				publickey, privatekey := GenerateSSHKeys(labrequest.ID.String())
				labrequest.PublicSSHKey = strings.TrimRight(string(publickey), "\n")

				AddSSHKeysToLabSecret(clientset, &labrequest, publickey, privatekey)
				AddOpenShiftVersionToLabSecret(clientset, &labrequest)
				GenerateInstallConfig(&labrequest)
				AddInstallConfigToLabSecret(clientset, &labrequest)
				CreateLabPullRequest(&labrequest, &CurrentLabRequestBranch, &issue, client, ctx, *org, *repo, *branch)
			} else {
				log.Printf("%v is an event being watched; action \"%v\" and state \"%v\" do not trigger.\n",
					r.Header.Get("X-GitHub-Event"),
					issue.Action,
					issue.Issue.State)
			}
		}
	})

	_ = http.ListenAndServe(":3000", nil)
}
