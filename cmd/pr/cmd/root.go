package cmd

import (
	"fmt"
	"github.com/hbk619/git-browse/cmd/pr/internal"
	"github.com/hbk619/git-browse/internal/github"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pr [number]",
	Args:  cobra.ExactArgs(1),
	Short: "Browse Github PR comments",
	Long:  `View comments from a PR one by one and reply to them`,
	Run: func(cmd *cobra.Command, args []string) {
		number, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error: Please provide a valid PR number")
			return
		}
		pr := internal.NewPRAction(github.NewPRClient(github.NewGHApi()))
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = pr.Init(number, verbose)
		if err != nil {
			fmt.Println(err)
		}
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
