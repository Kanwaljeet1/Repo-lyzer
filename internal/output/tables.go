package output

import (
	"fmt"
	"os"

	"github.com/agnivo988/Repo-lyzer/internal/github"

	"github.com/olekukonko/tablewriter"
)

func PrintRepo(r *github.Repo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Repository", "Stars", "Forks", "Open Issues"})
	table.Append([]string{
		r.FullName,
		fmt.Sprint(r.Stars),
		fmt.Sprint(r.Forks),
		fmt.Sprint(r.OpenIssues),
	})

	table.Render()
}
