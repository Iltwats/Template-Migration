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
)

// Method to raise a pull request
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

// Method to push the current local branch to remote
func pushTheBranch(name string) error {
	fmt.Println("Pushing the branch to remote..")
	pushCmd, err := GitCommand("push", "--set-upstream", "origin", name)
	if err != nil {
		return err
	}
	return PrepareCmd(pushCmd).Run()
}

// ApplyPatch applying the patch files
func ApplyPatch(filename string) error {
	patch, err := GitCommand("am", filename)
	if err != nil {
		return err
	}
	return PrepareCmd(patch).Run()
}

// CheckoutBranch Checkout Branch from master
func CheckoutBranch(branch string) error {
	configCmd, err := GitCommand("checkout", "-b", branch)
	if err != nil {
		return err
	}
	return PrepareCmd(configCmd).Run()
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

// GitCommand Misc Functions
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
