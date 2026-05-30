package output

import (
	"testing"

	"github.com/agnivo988/Repo-lyzer/internal/github"
)

func TestPrintRepo(t *testing.T) {
	repo := &github.Repo{
		FullName:   "owner/example",
		Stars:      120,
		Forks:      18,
		OpenIssues: 7,
	}

	// Regression guard: ensure table rendering path executes without panicking.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PrintRepo panicked: %v", r)
		}
	}()

	PrintRepo(repo)
}
