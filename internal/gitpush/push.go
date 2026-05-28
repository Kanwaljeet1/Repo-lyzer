package gitpush

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PushOptions holds options for the Git push operation.
type PushOptions struct {
	LocalPath string
	RepoOwner string
	RepoName  string
	CommitMsg string
	Token     string
	Branch    string
}

// PushRepo encapsulates staging, committing, remote URL configuration, and pushing to GitHub.
func PushRepo(opts PushOptions) error {
	// 1. Initialize git if .git directory does not exist
	gitDir := filepath.Join(opts.LocalPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmdInit := exec.Command("git", "init")
		cmdInit.Dir = opts.LocalPath
		if err := cmdInit.Run(); err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	// 2. Check current branch if none provided, or default to main
	branch := opts.Branch
	if branch == "" {
		cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmdBranch.Dir = opts.LocalPath
		branchBytes, err := cmdBranch.Output()
		branch = "main"
		if err == nil {
			b := strings.TrimSpace(string(branchBytes))
			if b != "" && b != "HEAD" {
				branch = b
			}
		}
	}

	// 3. Stage and commit files
	cmdStatus := exec.Command("git", "status", "--porcelain")
	cmdStatus.Dir = opts.LocalPath
	statusBytes, _ := cmdStatus.Output()
	if len(strings.TrimSpace(string(statusBytes))) > 0 {
		cmdAdd := exec.Command("git", "add", ".")
		cmdAdd.Dir = opts.LocalPath
		if err := cmdAdd.Run(); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		commitMsg := opts.CommitMsg
		if commitMsg == "" {
			commitMsg = "Push via Repo-lyzer"
		}
		cmdCommit := exec.Command("git", "commit", "-m", commitMsg)
		cmdCommit.Dir = opts.LocalPath
		_ = cmdCommit.Run() // ignore error if there is nothing to commit
	}

	// 4. Remote configuration
	cmdGetRemote := exec.Command("git", "remote", "get-url", "origin")
	cmdGetRemote.Dir = opts.LocalPath
	origBytes, errGet := cmdGetRemote.Output()
	hasOrigin := errGet == nil
	var originalURL string
	hasCustomRemote := false

	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", opts.RepoOwner, opts.RepoName)

	if hasOrigin {
		originalURL = strings.TrimSpace(string(origBytes))
		cmdSetRemote := exec.Command("git", "remote", "set-url", "origin", repoURL)
		cmdSetRemote.Dir = opts.LocalPath
		if err := cmdSetRemote.Run(); err != nil {
			return fmt.Errorf("failed to update remote origin URL to %s in %s: %w", repoURL, opts.LocalPath, err)
		}
		hasCustomRemote = true
	} else {
		cmdAddRemote := exec.Command("git", "remote", "add", "origin", repoURL)
		cmdAddRemote.Dir = opts.LocalPath
		if err := cmdAddRemote.Run(); err != nil {
			return fmt.Errorf("failed to add remote origin: %w", err)
		}
	}

	// Clean up remote configuration on function exit if push fails
	success := false
	defer func() {
		if !success {
			if !hasOrigin {
				cmdRemove := exec.Command("git", "remote", "remove", "origin")
				cmdRemove.Dir = opts.LocalPath
				_ = cmdRemove.Run()
			} else if hasCustomRemote && originalURL != "" {
				cmdRestore := exec.Command("git", "remote", "set-url", "origin", originalURL)
				cmdRestore.Dir = opts.LocalPath
				_ = cmdRestore.Run()
			}
		}
	}()

	// 5. Push to GitHub
	var cmdPush *exec.Cmd
	if opts.Token != "" {
		cmdPush = exec.Command("git", "-c", "credential.helper=!f() { echo password=$REPO_LYZER_TOKEN; }; f", "push", "-u", "origin", branch)
		cmdPush.Env = append(os.Environ(), "REPO_LYZER_TOKEN=" + opts.Token)
	} else {
		cmdPush = exec.Command("git", "push", "-u", "origin", branch)
	}
	cmdPush.Dir = opts.LocalPath
	if err := cmdPush.Run(); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	success = true

	return nil
}
