package cmd

import (
	"fmt"
	"os"

	"github.com/cli/cli/v2/git"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/hbk619/gh-peruse/cmd/pr/internal"
	"github.com/hbk619/gh-peruse/internal/filesystem"
	"github.com/hbk619/gh-peruse/internal/github"
	"github.com/hbk619/gh-peruse/internal/history"
	internal_os "github.com/hbk619/gh-peruse/internal/os"
	common "github.com/hbk619/gh-peruse/internal"
	"github.com/spf13/cobra"
)

var PRCmd = &cobra.Command{
	Use:   "pr [number]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Browse Github PR comments",
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
		clipboard := internal_os.NewClipboard()
		output := filesystem.NewStdOut()
		prompt := common.NewPrompt(os.Stdin, output)
		pr := internal.NewPRAction(prClient, historyService, output, clipboard, prompt)
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = pr.Init(args, verbose)
		if err != nil {
			fmt.Println(err)
			return
		}
		pr.Run()
	},
}

func Execute() {
	err := PRCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	PRCmd.AddCommand(CheckCommentCountCmd)
	PRCmd.Flags().BoolP("verbose", "v", false, "Verbose mode")
}
