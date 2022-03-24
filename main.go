package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cli/safeexec"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const APIEndpoint = "https://api.github.com/repos/"

type Release struct {
	TagName     string    `json:"tag_name"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
}

type Commits struct {
	SHA    string          `json:"sha"`
	NodeID string          `json:"node_id"`
	Url    string          `json:"html_url"`
	Parent []ParentCommits `json:"parents"`
}

// ParentCommits sub-structure of Commits
type ParentCommits struct {
	Sha     string `json:"sha"`
	Url     string `json:"url"`
	HtmlUrl string `json:"html_url"`
}

func main() {
	fmt.Println("Enter the stack repository in form of (User/RepoName)")
	var stackURL string
	_, err := fmt.Scanln(&stackURL)
	if err != nil {
		return
	}
	userInput := strings.Split(stackURL, "/")
	username := userInput[0]
	repoName := userInput[1]
	releaseURL := fmt.Sprintf(APIEndpoint+"%s/%s/releases", username, repoName)
	releaseData, err := getReleases(releaseURL)
	fmt.Println("Fetching all the release tags...")
	if err != nil {
		log.Fatalln(err)
	}
	// extract all the tags
	var tags []string
	for _, val := range releaseData {
		tags = append(tags, val.TagName)
	}
	//fmt.Println(tags)
	tagSelectedByUser := tags[0]
	//userRepoConsumedTag := tags[2] // TODO fetch from API
	isUserRepoStackConsumed := true // TODO fetch from API

	if isUserRepoStackConsumed {
		commitsUrl := fmt.Sprintf(APIEndpoint+"%s/%s/commits/%s", username, repoName, tagSelectedByUser)
		commitsResp, comErr := getCommits(commitsUrl)
		fmt.Println("Fetching all the commits for each tags...")
		if comErr != nil {
			panic(comErr)
		}
		parents := commitsResp.Parent
		var parentUrls []string
		for _, parent := range parents {
			parentUrls = append(parentUrls, parent.HtmlUrl)
		}

		var patchFileUrls []string
		patchFileUrls = append(patchFileUrls, commitsResp.Url)
		for _, url := range parentUrls {
			patchFileUrls = append(patchFileUrls, url)
		}

		isPatchFileDownloaded := savePatchFile(patchFileUrls, tagSelectedByUser)

		if isPatchFileDownloaded {
			applyPatchFile(tagSelectedByUser, len(patchFileUrls))
		}
	}

}

func applyPatchFile(tag string, indices int) {
	branchName := "patch-apply"
	err := CheckoutBranch(branchName)
	if err != nil {
		panic("Checkout " + err.Error())
	}
	var patchFileNames []string
	for i := 0; i < indices; i++ {
		name := fmt.Sprintf("%s-%d.patch", tag, i)
		patchFileNames = append(patchFileNames, name)
	}
	for _, name := range patchFileNames {
		patchErr := ApplyPatch(name)
		if patchErr != nil {
			panic("Patch " + patchErr.Error())
			return
		}
	}
	fmt.Println("All the patch files applied successfully.")
	isCacheCleared := DeleteCache(patchFileNames)
	if isCacheCleared {
		pushErr := pushTheBranch(branchName)
		if pushErr != nil {
			log.Fatalln("Error while pushing ", pushErr)
		}
		fmt.Println("Successfully pushed to remote")
		raiseAPullRequest()
	}

}

func raiseAPullRequest() {
	fmt.Println("Creating a pull request...")
	prTitle := "Migration-patch"
	prBody := "PR to migrate to latest stack version"
	cmd, err := exec.Command("gh", "pr", "create", "--title", prTitle, "--body", prBody).Output()
	if err != nil {
		fmt.Println("Error while creating a PR", err)
	}
	fmt.Println("To complete the merge, merge this PR by going to the following link: ", string(cmd))
}

func pushTheBranch(name string) error {
	fmt.Println("Pushing the branch to remote..")
	pushCmd, err := GitCommand("push", "--set-upstream", "origin", name)
	if err != nil {
		return err
	}
	return PrepareCmd(pushCmd).Run()
}

func DeleteCache(names []string) bool {
	for _, name := range names {
		_, err := exec.Command("rm", "-rf", name).Output()
		if err != nil {
			fmt.Println("Error while deleting", err)
			return false
		}
	}
	fmt.Println("Patch files cache removed ")
	return true
}

func ApplyPatch(filename string) error {
	patch, err := GitCommand("am", filename)
	if err != nil {
		return err
	}
	return PrepareCmd(patch).Run()
}

func CheckoutBranch(branch string) error {
	configCmd, err := GitCommand("checkout", "-b", branch)
	if err != nil {
		return err
	}
	return PrepareCmd(configCmd).Run()
}
func GitCommand(args ...string) (*exec.Cmd, error) {
	gitExe, err := safeexec.LookPath("git")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			programName := "git"
			if runtime.GOOS == "windows" {
				programName = "Git for Windows"
			}
			return nil, &NotInstalled{
				message: fmt.Sprintf("unable to find git executable in PATH; please install %s before retrying", programName),
				error:   err,
			}
		}
		return nil, err
	}
	return exec.Command(gitExe, args...), nil
}

type NotInstalled struct {
	message string
	error
}

func (e *NotInstalled) Error() string {
	return e.message
}

// Fetch all the release tags available for stack repository
func getReleases(url string) ([]Release, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	var release []Release
	parseError := json.NewDecoder(resp.Body).Decode(&release)
	defer resp.Body.Close()
	return release, parseError

}

// Fetch all the commits to its corresponding tags available for stack repository
func getCommits(url string) (Commits, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	var commits Commits
	parseError := json.NewDecoder(resp.Body).Decode(&commits)

	defer resp.Body.Close()
	return commits, parseError

}

// Method to indent the JSON and view
func printIndentedJSON(ert interface{}) {
	data, err := json.MarshalIndent(ert, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

// Method to download and save patch file
func savePatchFile(urls []string, tag string) bool {
	fmt.Println("Downloading Patch files...")
	var fileLen = 0
	for i, url := range urls {
		i = len(urls) - 1 - i
		name := fmt.Sprintf("%s-%d", tag, i)
		out, _ := os.Create(name + ".patch")
		// timeout if it takes more than 10 secs
		client := http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url + ".patch")
		if err != nil {
			log.Fatalln("Timeout", err.Error())
		}
		_, _ = io.Copy(out, resp.Body)
		fmt.Printf("Download complete for patch file -%d\n", i)
		fileLen++
		resp.Body.Close()
		out.Close()
	}
	return fileLen == len(urls)
}

// Runnable is typically an exec.Cmd or its stub in tests
type Runnable interface {
	Output() ([]byte, error)
	Run() error
}

// PrepareCmd extends exec.Cmd with extra error reporting features and provides a
// hook to stub command execution in tests
var PrepareCmd = func(cmd *exec.Cmd) Runnable {
	return &cmdWithStderr{cmd}
}

// cmdWithStderr augments exec.Cmd by adding stderr to the error message
type cmdWithStderr struct {
	*exec.Cmd
}

func (c cmdWithStderr) Output() ([]byte, error) {
	if os.Getenv("DEBUG") != "" {
		_ = printArgs(os.Stderr, c.Cmd.Args)
	}
	if c.Cmd.Stderr != nil {
		return c.Cmd.Output()
	}
	errStream := &bytes.Buffer{}
	c.Cmd.Stderr = errStream
	out, err := c.Cmd.Output()
	if err != nil {
		err = &CmdError{errStream, c.Cmd.Args, err}
	}
	return out, err
}

func (c cmdWithStderr) Run() error {
	if os.Getenv("DEBUG") != "" {
		_ = printArgs(os.Stderr, c.Cmd.Args)
	}
	if c.Cmd.Stderr != nil {
		return c.Cmd.Run()
	}
	errStream := &bytes.Buffer{}
	c.Cmd.Stderr = errStream
	err := c.Cmd.Run()
	if err != nil {
		err = &CmdError{errStream, c.Cmd.Args, err}
	}
	return err
}

// CmdError provides more visibility into why an exec.Cmd had failed
type CmdError struct {
	Stderr *bytes.Buffer
	Args   []string
	Err    error
}

func (e CmdError) Error() string {
	msg := e.Stderr.String()
	if msg != "" && !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	return fmt.Sprintf("%s%s: %s", msg, e.Args[0], e.Err)
}

func printArgs(w io.Writer, args []string) error {
	if len(args) > 0 {
		// print commands, but omit the full path to an executable
		args = append([]string{filepath.Base(args[0])}, args[1:]...)
	}
	_, err := fmt.Fprintf(w, "%v\n", args)
	return err
}
