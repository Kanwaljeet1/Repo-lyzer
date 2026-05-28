package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agnivo988/Repo-lyzer/internal/config"
	"github.com/agnivo988/Repo-lyzer/internal/progress"
	"github.com/spf13/cobra"
)

var (
	pushToken  string
	pushMsg    string
	pushBranch string
)

func cleanAndResolvePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	// Fix common user typos where they omit the leading slash on absolute Unix paths
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, ".") && !strings.HasPrefix(path, "~") {
		slashed := "/" + path
		if stat, err := os.Stat(slashed); err == nil && stat.IsDir() {
			path = slashed
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return filepath.Clean(absPath), nil
}

var pushCmd = &cobra.Command{
	Use:   "push <local-folder> <owner/repo>",
	Short: "Push a local folder or repository to GitHub",
	Long: `Initialize, commit, and push a local folder or existing repository 
to a remote GitHub repository.

If the local folder is not already a Git repository, it will be initialized 
automatically. Any untracked/modified files will be staged and committed 
before pushing.

Authentication:
Uses your configured GitHub Personal Access Token. Alternatively, you can 
provide a token via the --token (-t) flag or set the GITHUB_TOKEN environment 
variable.

Examples:
  # Push local folder to remote repo
  repo-lyzer push /path/to/project octocat/my-new-repo

  # With custom commit message and branch
  repo-lyzer push ./my-app octocat/my-app -m "Deploy v1.0" -b production`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		localPath := args[0]
		repoName := args[1]

		parts := strings.Split(repoName, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid repository: must be in owner/repo format")
		}
		owner := parts[0]
		repo := parts[1]

		// Resolve and clean local folder path
		resolvedPath, err := cleanAndResolvePath(localPath)
		if err != nil {
			return err
		}

		// Check if local folder exists
		if stat, err := os.Stat(resolvedPath); os.IsNotExist(err) {
			return fmt.Errorf("local folder does not exist: %s", resolvedPath)
		} else if !stat.IsDir() {
			return fmt.Errorf("path is not a directory: %s", resolvedPath)
		}
		localPath = resolvedPath

		spinner := progress.NewSpinner()

		// 1. Git initialization if not already a git repo
		gitDir := filepath.Join(localPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			spinner.Start("Initializing Git repository...")
			cmdInit := exec.Command("git", "init")
			cmdInit.Dir = localPath
			if err := cmdInit.Run(); err != nil {
				spinner.Stop()
				return fmt.Errorf("failed to initialize git: %w", err)
			}
			spinner.StopWithMessage("Initialized empty Git repository")
		}

		// 2. Commit any uncommitted changes
		cmdStatus := exec.Command("git", "status", "--porcelain")
		cmdStatus.Dir = localPath
		statusBytes, _ := cmdStatus.Output()
		if len(strings.TrimSpace(string(statusBytes))) > 0 {
			spinner.Start("Staging and committing files...")
			cmdAdd := exec.Command("git", "add", ".")
			cmdAdd.Dir = localPath
			if err := cmdAdd.Run(); err != nil {
				spinner.Stop()
				return fmt.Errorf("failed to stage files: %w", err)
			}

			cmdCommit := exec.Command("git", "commit", "-m", pushMsg)
			cmdCommit.Dir = localPath
			_ = cmdCommit.Run() // ignore error if there is nothing to commit
			spinner.StopWithMessage("Committed files")
		}

		// 3. Remote configuration
		spinner.Start("Configuring remote repository...")
		settings, err := config.LoadSettings()
		token := pushToken
		if token == "" && err == nil {
			token = settings.GitHubToken
		}
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		repoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
		if token != "" {
			repoURL = fmt.Sprintf("https://%s@github.com/%s/%s.git", token, owner, repo)
		}

		cmdGetRemote := exec.Command("git", "remote", "get-url", "origin")
		cmdGetRemote.Dir = localPath
		origBytes, errGet := cmdGetRemote.Output()
		hasOrigin := errGet == nil
		var originalURL string
		hasCustomRemote := false

		if hasOrigin {
			originalURL = strings.TrimSpace(string(origBytes))
			cmdSetRemote := exec.Command("git", "remote", "set-url", "origin", repoURL)
			cmdSetRemote.Dir = localPath
			if err := cmdSetRemote.Run(); err == nil {
				hasCustomRemote = true
			}
		} else {
			cmdAddRemote := exec.Command("git", "remote", "add", "origin", repoURL)
			cmdAddRemote.Dir = localPath
			if err := cmdAddRemote.Run(); err != nil {
				spinner.Stop()
				return fmt.Errorf("failed to add remote origin: %w", err)
			}
		}

		// Clean up remote URL on function exit
		defer func() {
			if hasCustomRemote && originalURL != "" {
				cmdRestore := exec.Command("git", "remote", "set-url", "origin", originalURL)
				cmdRestore.Dir = localPath
				_ = cmdRestore.Run()
			}
		}()
		spinner.StopWithMessage("Configured remote origin")

		// 4. Pushing code
		spinner.Start(fmt.Sprintf("Pushing to GitHub (%s)...", pushBranch))
		cmdPush := exec.Command("git", "push", "-u", "origin", pushBranch)
		cmdPush.Dir = localPath
		if err := cmdPush.Run(); err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to push: %w", err)
		}
		spinner.StopWithMessage(fmt.Sprintf("✓ Pushed code successfully to %s/%s branch %s!", owner, repo, pushBranch))

		return nil
	},
}

func init() {
	pushCmd.Flags().StringVarP(&pushToken, "token", "t", "", "GitHub Personal Access Token")
	pushCmd.Flags().StringVarP(&pushMsg, "message", "m", "Push via Repo-lyzer", "Commit message")
	pushCmd.Flags().StringVarP(&pushBranch, "branch", "b", "main", "Target branch")

	rootCmd.AddCommand(pushCmd)
}
