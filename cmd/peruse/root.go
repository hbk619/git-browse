package peruse

import (
	"github.com/hbk619/gh-peruse/cmd/pr/cmd"
	"github.com/spf13/cobra"
	"os"
)

var PeruseCmd = &cobra.Command{
	Use:   "peruse",
	Short: "Look at things in Github",
	Long:  "Look at things in Github, one by one",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	PeruseCmd.AddCommand(cmd.PRCmd)
}

func Execute() {
	err := PeruseCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
