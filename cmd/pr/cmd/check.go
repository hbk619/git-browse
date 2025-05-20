package cmd

import (
	"fmt"
	"os"

	"github.com/cli/cli/v2/git"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/hbk619/gh-peruse/cmd/pr/internal/new_comments"

	"github.com/hbk619/gh-peruse/internal/filesystem"
	"github.com/hbk619/gh-peruse/internal/github"
	"github.com/hbk619/gh-peruse/internal/history"
	"github.com/hbk619/gh-peruse/internal/notifications"
	"github.com/spf13/cobra"
)

var CheckCommentCountCmd = &cobra.Command{
	Use:   "check",
	Args:  cobra.NoArgs,
	Short: "Check and notify of new commands",
	Long:  `View comments from a PR one by one and reply to them`,
	Run: func(cmd *cobra.Command, args []string) {
		historyService, err := history.NewHistoryService(os.Getenv("HOME"), filesystem.NewFS())
		if err != nil {
			fmt.Println(err)
			return
		}
		graphQlClient, err := api.DefaultGraphQLClient()
		if err != nil {
			fmt.Println(err)
			return
		}
		gitClient := &git.Client{}

		prClient := github.NewPRClient(graphQlClient, gitClient)
		if err != nil {
			fmt.Println(err)
			return
		}
		notify, err := cmd.Flags().GetBool("notify")
		if err != nil {
			fmt.Println(err)
			return
		}
		var output filesystem.Output
		if notify {
			output = notifications.NewNotifier()
		} else {
			output = filesystem.NewStdOut()
		}
		err = new_comments.CheckForNewComments(prClient, historyService, output)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	CheckCommentCountCmd.Flags().BoolP("notify", "n", false, "Show notification for new comments")
}
