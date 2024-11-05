package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/tommyshem/ogi/cmd/issue"
	"golang.org/x/oauth2"
)

var fetchState string

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetches all of the issues for the specified repo.",
	Long: `Fetches all of the issues for the specified repo.
This will clear any existing issues stored locally, pull down
everything from GitHub and store it locally for offline use.

The first time you run this command you should run it like such:

$ ghi fetch owner/repo

Subsequent calls will not need the "owner/repo".

If you are going to be calling a private repo you will need to
set the ENV var "GITHUB_TOKEN" with a GitHub Personal Access Token.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			config.SetFromArgs(args)
		}

		err := db.Clear()
		if err != nil {
			log.Fatal(err)
		}

		more := true
		count := 0
		opts := &github.IssueListByRepoOptions{State: fetchState}

		spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		spin.Suffix = fmt.Sprintf(" Fetching %s Issues (This could take a while depending on the number of issues and comments you have!)", fetchState)
		spin.Start()

		client := newClient()
		for more {
			issues, resp, err := client.Issues.ListByRepo(context.Background(), db.Owner, db.Repo, opts)
			if err != nil {
				fmt.Printf("\n%s", err.Error())
				if resp != nil && resp.StatusCode == 401 {
					fmt.Println(`Couldn't access this repo! Try setting a GitHub Personal Token.

This token can be set as an environment variable "GITHUB_TOKEN".`)
				}
				os.Exit(2)
			}
			count += len(issues)

			Wait(len(issues), func(i int) {
				issue := &issue.Issue{Issue: *issues[i], Comments: []*github.IssueComment{}}
				comments, _, err := client.Issues.ListComments(context.Background(), db.Owner, db.Repo, *issue.Number, &github.IssueListCommentsOptions{})
				if err != nil {
					log.Fatal(err)
				}
				issue.Comments = comments
				db.Save(*issue)
			})
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		if err != nil {
			log.Fatal(err)
		}
		config.Save()
		spin.Stop()
		fmt.Printf("\nFetched %d issues for %s/%s\n", count, db.Owner, db.Repo)
	},
}

func Wait(length int, block func(index int)) {
	var w sync.WaitGroup
	w.Add(length)
	for i := 0; i < length; i++ {
		go func(w *sync.WaitGroup, index int) {
			block(index)
			w.Done()
		}(&w, i)
	}
	w.Wait()
}

// newClient returns a new github.Client. If the GITHUB_TOKEN environment
// variable is set, a client with that token is created. Otherwise a client
// with no special auth is created.
func newClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) > 0 {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		return github.NewClient(tc)
	}
	return github.NewClient(&http.Client{})
}

// init registers the fetch command with the root command and adds a flag to
// specify the state of issues to fetch.
func init() {
	RootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringVarP(&fetchState, "state", "s", "all", "Fetch issues by their state <all, closed, open>")
}
