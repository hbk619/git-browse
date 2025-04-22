package internal

import (
	"errors"
	"fmt"
	"github.com/hbk619/git-browse/internal"
	"github.com/hbk619/git-browse/internal/filesystem"
	"github.com/hbk619/git-browse/internal/git"
	"github.com/hbk619/git-browse/internal/github"
	"github.com/hbk619/git-browse/internal/history"
	"os"
	"strings"
)

type PRAction struct {
	Index               int
	Repo                *git.Repo
	Results             []git.Comment
	PrintedPathLastTime bool
	MaxIndex            int
	LastFullPath        string
	HelpText            string
	State               git.State
	client              github.PullRequestClient
	history             history.Storage
	output              filesystem.Output
	internal.Interactive
}

func NewPRAction(client github.PullRequestClient, history history.Storage, output filesystem.Output) *PRAction {
	return &PRAction{
		Repo:                &git.Repo{},
		PrintedPathLastTime: true,
		LastFullPath:        "",
		HelpText:            "Type c to comment",
		client:              client,
		history:             history,
		output:              output,
	}
}

func (pr *PRAction) Init(prNumber int, verbose bool) error {
	repoDetails, err := pr.client.GetRepoDetails()
	if err != nil {
		return err
	}
	pr.Repo.Owner = repoDetails.Owner
	pr.Repo.Name = repoDetails.Name
	pr.Repo.PRNumber = prNumber

	prDetails, err := pr.client.GetPRDetails(pr.Repo, verbose)
	if err != nil {
		return err
	}
	pr.Results = prDetails.Comments
	pr.State = prDetails.State

	if verbose {
		commitComments, err := pr.client.GetCommitComments(pr.Repo.Owner, pr.Repo.Name, prNumber)
		if err != nil {
			return err
		}
		pr.Results = append(pr.Results, commitComments...)
	}

	if verbose {
		pr.PrintState()
	}

	commentCount := len(pr.Results)
	prHistory, err := pr.history.Load()
	if err != nil {
		pr.output.Print(fmt.Sprintf("Warning failed to load comments to history: %s", err.Error()))
	}

	if err == nil {
		existingPrHistory := prHistory.Prs[prNumber]
		if existingPrHistory.CommentCount != commentCount {
			pr.output.Print("New comments ahead!")
		}

		existingPrHistory.CommentCount = commentCount
		prHistory.Prs[prNumber] = existingPrHistory
		err = pr.history.Save(prHistory)
		if err != nil {
			pr.output.Print(fmt.Sprintf("Warning failed to save comments to history: %s", err.Error()))
		}
	}

	if commentCount == 0 {
		return errors.New("no comments found")
	}

	pr.Interactive.MaxIndex = commentCount - 1
	pr.LastFullPath = pr.Results[0].FileDetails.FullPath
	pr.Print()
	return nil
}

func (pr *PRAction) Run() {
	for {
		result := internal.StringPrompt("n to go to the next result, p for previous, r to repeat or q to quit")
		switch result {
		case "n":
			pr.Interactive.Next(pr.Print)
		case "p":
			pr.Interactive.Previous(pr.Print)
		case "r":
			pr.Interactive.Repeat(pr.Print)
		case "q":
			os.Exit(0)
		default:
			pr.output.Print("Invalid choice")
		}

	}
}

func (pr *PRAction) Print() {
	current := pr.Results[pr.Interactive.Index]
	if current.Thread.IsResolved {
		pr.output.Print("This comment is resolved")
	}
	if current.Outdated {
		pr.output.Print("This comment is outdated")
	}
	pr.output.Print(current.Author.Login)
	pr.output.Print(current.Body)
}

func (pr *PRAction) PrintState() {
	pr.output.Print(pr.State.MergeStatus)
	pr.output.Print(pr.State.ConflictStatus)
	for reviewState, names := range pr.State.Reviews {
		pr.output.Print(fmt.Sprintf("%s %s", reviewState, strings.Join(names, " ")))
	}
	for _, status := range pr.State.Statuses {
		pr.output.Print(fmt.Sprintf("Check %s %s", status.Name, status.Conclusion))
	}
}
