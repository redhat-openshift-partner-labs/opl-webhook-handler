package libs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
	whgh "gopkg.in/go-playground/webhooks.v5/github"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	. "k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"strconv"
	"time"
)

func GithubAuthenticate() (*github.Client, context.Context) {
	accesstoken := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accesstoken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, ctx
}

func K8sAuthenticate() *kubernetes.Clientset {
	// create k8s client
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("OPENSHIFT_KUBECONFIG"))
	ErrorCheck("The kubeconfig could not be loaded", err)
	clientset, err := kubernetes.NewForConfig(cfg)

	return clientset
}

func DefaultClientK8sAuthenticate() (*rest.Config, error) {
	cfg, err := clientcmd.LoadFromFile(os.Getenv("OPENSHIFT_KUBECONFIG"))
	ErrorCheck("The kubeconfig could not be loaded", err)
	client := clientcmd.NewDefaultClientConfig(*cfg, &clientcmd.ConfigOverrides{})

	return client.ClientConfig()
}

func DynamicClientK8sAuthenticate() (Interface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("OPENSHIFT_KUBECONFIG"))
	ErrorCheck("The kubeconfig could not be loaded", err)
	client, err := NewForConfig(cfg)

	return client, err
}

func ErrorCheck(message string, err error) (ok bool) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
	return true
}

// function that removes completed, expired, or deleted lab github branches
func DeleteBranch(request *LabRequestBranch, client *github.Client, ctx context.Context) {
	_, err := client.Git.DeleteRef(ctx, "opdev", "lab-requests", "heads/" + request.Lab)
	ErrorCheck("Unable to delete branch", err)
}

func CreateBranch(labRequest *LabRequest, requestBranch *LabRequestBranch, client *github.Client, ctx context.Context) *github.Reference {
	requestJSON, err := json.Marshal(&labRequest)
	ErrorCheck("Unable to marshal lab request struct", err)

	err = ioutil.WriteFile("/tmp/"+labRequest.ID.String()+".json", requestJSON, 0644)
	ErrorCheck("Unable to create json file", err)

	branch := &github.Reference{Ref: github.String("refs/heads/" + labRequest.ID.String()), Object: &github.GitObject{SHA: &requestBranch.Base}}
	ref, _, err := client.Git.CreateRef(ctx, "opdev", "opl-requests", branch)
	ErrorCheck("Unable to create the branch", err)

	return ref
}

func CreateLabPullRequest(labRequest *LabRequest, requestBranch *LabRequestBranch, issue *whgh.IssuesPayload, client *github.Client, ctx context.Context) {
	ref := CreateBranch(labRequest, requestBranch, client, ctx)

	requestJSON, err := json.Marshal(&labRequest)

	var entry []*github.TreeEntry
	entry = append(entry, &github.TreeEntry{
		Path:    github.String("labs/" + labRequest.ID.String() + ".json"),
		Type:    github.String("blob"),
		Content: github.String(string(requestJSON)),
		Mode:    github.String("100644"),
	})

	tree, _, err := client.Git.CreateTree(ctx, "opdev", "opl-requests", requestBranch.Lab, entry)

	// Add the request file to the new branch
	date := time.Now()
	author := &github.CommitAuthor{
		Date: &date, Name: github.String("Lifecycle Engineering"),
		Email: github.String("sd-ecosystem@redhat.com"),
	}

	parent, _, err := client.Repositories.GetCommit(ctx, "opdev", "opl-requests", requestBranch.Lab)
	ErrorCheck("Unable to get commit", err)
	parent.Commit.SHA = parent.SHA
	commitParent := parent.GetCommit()

	commit := github.Commit{
		Author:  author,
		Message: github.String("Triaged Lab Request: " + labRequest.ID.String()),
		Tree:    tree,
		Parents: []*github.Commit{commitParent},
	}

	commitData, _, err := client.Git.CreateCommit(ctx, "opdev", "opl-requests", &commit)
	ErrorCheck("Commit creation failed", err)
	ref.Object.SHA = commitData.SHA
	_, _, err = client.Git.UpdateRef(ctx, "opdev", "opl-requests", ref, false)

	// Create the pull request for the new lab request
	requestPR := &github.NewPullRequest{
		Title:               github.String("Lab Request: " + labRequest.ID.String()),
		Head:                github.String("opdev:" + labRequest.ID.String()),
		Base:                github.String("master"),
		Body:                github.String("associated with issue #" + strconv.Itoa(int(issue.Issue.Number))),
		MaintainerCanModify: github.Bool(false),
	}

	_, _, err = client.PullRequests.Create(ctx, "opdev", "opl-requests", requestPR)
	if ErrorCheck("Unable to create pull request", err) {
		DeleteIssue(issue, client, ctx)
	}
}

func DeleteIssue(issue *whgh.IssuesPayload, client *github.Client, ctx context.Context) {
	//targetIssue, _, err := client.Issues.Get(ctx, "opdev", "lab-requests", int(issue.Issue.Number))
	issueState := &github.IssueRequest{State: github.String("closed")}
	_, _, err := client.Issues.Edit(ctx, "opdev", "opl-requests", int(issue.Issue.Number), issueState)
	ErrorCheck("Unable to close issue", err)
}

func CreateLabSecret(clientset *kubernetes.Clientset, labRequest *LabRequest) {
	labSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: labRequest.ID.String(),
		},
		Type: "Opaque",
	}

	_, err := clientset.CoreV1().Secrets("hive").Create(context.Background(), &labSecret, metav1.CreateOptions{})
	ErrorCheck("Unable to create secret", err)
}

func AddSSHKeysToLabSecret(clientset *kubernetes.Clientset, labRequest *LabRequest, publickey []byte, privatekey []byte) {
	pubkey := base64.StdEncoding.EncodeToString(publickey)
	prikey := base64.StdEncoding.EncodeToString(privatekey)
	data := "{\"data\":{\"ssh-publickey\": \"" + pubkey + "\", \"ssh-privatekey\": \"" + prikey + "\"}}"

	_, err := clientset.CoreV1().Secrets("hive").Patch(context.Background(),
		labRequest.ID.String(),
		types.StrategicMergePatchType,
		[]byte(data),
		metav1.PatchOptions{})
	ErrorCheck("Unable to patch secret " + labRequest.ID.String() + ": ", err)
}

func AddInstallConfigToLabSecret(clientset *kubernetes.Clientset, labRequest *LabRequest) {
  installconfig := base64.StdEncoding.EncodeToString(GenerateInstallConfig(labRequest))

	data := "{\"data\":{\"install-config.yaml\": \"" + installconfig + "\"}}"
	_, err := clientset.CoreV1().Secrets("hive").Patch(context.Background(),
		labRequest.ID.String(),
		types.StrategicMergePatchType,
		[]byte(data),
		metav1.PatchOptions{})
	ErrorCheck("Unable to patch secret " + labRequest.ID.String() + ": ", err)
}

func AddOpenShiftVersionToLabSecret(clientset *kubernetes.Clientset, labRequest *LabRequest) {
	openshift := base64.StdEncoding.EncodeToString([]byte(labRequest.OpenShiftVersion))

	data := "{\"data\":{\"openshift\": \"" + openshift + "\"}}"
	_, err := clientset.CoreV1().Secrets("hive").Patch(context.Background(),
		labRequest.ID.String(),
		types.StrategicMergePatchType,
		[]byte(data),
		metav1.PatchOptions{})
	ErrorCheck("Unable to add OpenShift Version to secret " + labRequest.ID.String() + ": ", err)
}

func RemoveArtifacts(artifacts []string) {
	for _, artifact := range artifacts {
		err := os.Remove(artifact)
		ErrorCheck("Unable to remove file: %v", err)
	}
}