package internal

import (
	"fmt"
	"github.com/hbk619/git-browse/internal"
	"github.com/hbk619/git-browse/internal/git"
	"github.com/hbk619/git-browse/internal/github"
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
	internal.Interactive
}

func NewPRAction(client github.PullRequestClient) *PRAction {
	return &PRAction{
		Repo:                &git.Repo{},
		PrintedPathLastTime: true,
		LastFullPath:        "",
		HelpText:            "Type c to comment",
		client:              client,
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

	prDetails, err := pr.client.GetMainPRDetails(prNumber, verbose)
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

	if len(pr.Results) == 0 {
		fmt.Println("No comments found")
		os.Exit(1)
	}

	pr.Interactive.MaxIndex = len(pr.Results) - 1
	pr.LastFullPath = pr.Results[0].FileDetails.FullPath
	pr.Print()
	pr.R()
	return nil
}

func (pr *PRAction) R() {
	for {
		result := internal.StringPrompt("n to go to the next result, p for previous, r to repeat or q to quit")
		switch result {
		case "n":
			pr.Interactive.N(pr.Print)
		case "p":
			pr.Interactive.P(pr.Print)
		case "r":
			pr.Interactive.R(pr.Print)
		case "q":
			os.Exit(0)
		default:
			fmt.Println("Invalid choice")
		}

	}
}

func (pr *PRAction) Print() {
	current := pr.Results[pr.Interactive.Index]
	if current.Thread.IsResolved {
		fmt.Println("This comment is resolved")
	}
	if current.Outdated {
		fmt.Println("This comment is outdated")
	}
	fmt.Println(current.Author.Login)
	fmt.Println(current.Body)
}

func (pr *PRAction) PrintState() {
	fmt.Println(pr.State.MergeStatus)
	fmt.Println(pr.State.ConflictStatus)
	for reviewState, names := range pr.State.Reviews {
		fmt.Printf("%s %s\n", reviewState, strings.Join(names, " "))
	}
	for _, status := range pr.State.Statuses {
		fmt.Printf("Check %s %s\n", status.Name, status.Conclusion)
	}
}
