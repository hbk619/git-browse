package cmd

import (
	"fmt"
	"github.com/cli/cli/v2/git"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/hbk619/git-browse/cmd/pr/internal"
	"github.com/hbk619/git-browse/internal/filesystem"
	"github.com/hbk619/git-browse/internal/github"
	"github.com/hbk619/git-browse/internal/history"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
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
		pr := internal.NewPRAction(prClient, historyService, filesystem.NewStdOut())
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("verbose", "v", false, "Verbose mode")
}
