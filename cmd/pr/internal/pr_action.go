package internal

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hbk619/gh-peruse/internal"
	"github.com/hbk619/gh-peruse/internal/filesystem"
	"github.com/hbk619/gh-peruse/internal/git"
	"github.com/hbk619/gh-peruse/internal/github"
	"github.com/hbk619/gh-peruse/internal/history"
	internal_os "github.com/hbk619/gh-peruse/internal/os"
)

type PRAction struct {
	Id                  string
	Repo                *git.Repo
	Results             []git.Comment
	PrintedPathLastTime bool
	LastFullPath        string
	HelpText            string
	State               git.State
	client              github.PullRequestClient
	history             history.Storage
	output              filesystem.Output
	clipboard           internal_os.Clippy
	prompt              internal.Prompter
	internal.Interactive
}

func NewPRAction(client github.PullRequestClient, history history.Storage, output filesystem.Output, clipboard internal_os.Clippy) *PRAction {
	return &PRAction{
		Repo:                &git.Repo{},
		PrintedPathLastTime: true,
		LastFullPath:        "",
		HelpText:            "Type c to comment",
		client:              client,
		history:             history,
		output:              output,
		clipboard:           clipboard,
		prompt:              *internal.NewPrompt(os.Stdin, output),
	}
}

func (pr *PRAction) Init(args []string, verbose bool) error {
	repoDetails, err := pr.client.GetRepoDetails()
	if err != nil {
		return err
	}
	pr.Repo.Owner = repoDetails.Owner
	pr.Repo.Name = repoDetails.Name

	pr.Repo.PRNumber, err = pr.getPRNumber(args)
	if err != nil {
		return err
	}

	prDetails, err := pr.client.GetPRDetails(pr.Repo, verbose)
	if err != nil {
		return err
	}
	pr.Results = prDetails.Comments
	pr.State = prDetails.State
	pr.Id = prDetails.Id
	if verbose {
		pr.PrintState()
	}

	commentCount := len(pr.Results)
	pr.updateHistory(pr.Repo.PRNumber, commentCount)

	if commentCount == 0 {
		return errors.New("no comments found")
	}

	pr.Interactive.MaxIndex = commentCount - 1
	pr.Print()
	return nil
}

func (pr *PRAction) getPRNumber(args []string) (int, error) {
	if len(args) > 0 {
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return 0, fmt.Errorf("please provide a valid PR number")
		}
		return number, nil
	} else {
		return pr.client.DetectCurrentPR(pr.Repo)
	}
}

func (pr *PRAction) updateHistory(prNumber int, commentCount int) {
	prHistory, err := pr.history.Load()
	if err != nil {
		_ = pr.output.Println(fmt.Sprintf("Warning failed to load comments to history: %s", err.Error()))
		return
	}

	existingPrHistory := prHistory.Prs[prNumber]
	if existingPrHistory.CommentCount != commentCount {
		_ = pr.output.Println("New comments ahead!")
	}

	existingPrHistory.CommentCount = commentCount
	prHistory.Prs[prNumber] = existingPrHistory
	err = pr.history.Save(prHistory)
	if err != nil {
		_ = pr.output.Println(fmt.Sprintf("Warning failed to save comments to history: %s", err.Error()))
	}
}

func (pr *PRAction) Reply(contents string) {
	err := pr.client.Reply(contents, &pr.Results[pr.Interactive.Index], pr.Id)
	if err != nil {
		_ = pr.output.Println(fmt.Sprintf("Warning failed to comment: %s", err.Error()))
	} else {
		_ = pr.output.Println("Posted comment")
	}
}

func (pr *PRAction) Resolve() {
	err := pr.client.Resolve(&pr.Results[pr.Interactive.Index])
	if err != nil {
		_ = pr.output.Println(fmt.Sprintf("Warning failed to resolve thread: %s", err.Error()))
	} else {
		_ = pr.output.Println("Conversation resolved")
	}
}

func (pr *PRAction) Run() {
	for {
		prompt := "n to go to the next result, p for previous, r to repeat, x to copy or q to quit"
		currentComment := pr.Results[pr.Interactive.Index]
		pr.LastFullPath = currentComment.File.FullPath
		if currentComment.Thread.ID != "" && !currentComment.Thread.IsResolved {
			prompt += ", res to resolve"
		}
		if currentComment.Thread.IsResolved || currentComment.Outdated {
			prompt += ", e to expand"
		}
		result := pr.prompt.String(prompt)
		switch result {
		case "n":
			pr.Interactive.Next(pr.Print)
		case "p":
			pr.Interactive.Previous(pr.Print)
		case "r":
			pr.Interactive.Repeat(pr.Print)
		case "e":
			pr.LastFullPath = ""
			pr.printContents(currentComment)
		case "res":
			pr.Resolve()
		case "c":
			comment := pr.prompt.String("Type comment and press enter")
			pr.Reply(comment)
		case "x":
			err := pr.clipboard.Write(currentComment.Body)
			if err != nil {
				pr.output.Println(err.Error())
			}
		case "q":
			os.Exit(0)
		default:
			_ = pr.output.Println("Invalid choice")
		}

	}
}

func (pr *PRAction) Print() {
	current := pr.Results[pr.Interactive.Index]
	if current.Thread.IsResolved {
		_ = pr.output.Println("This comment is resolved")
		return
	}
	if current.Outdated {
		_ = pr.output.Println("This comment is outdated")
		return
	}
	pr.printContents(current)
}

func (pr *PRAction) printContents(current git.Comment) {
	if pr.LastFullPath != current.File.FullPath {
		_ = pr.output.Println(current.File.FileName)
		if current.File.Path != "" {
			_ = pr.output.Println(current.File.Path)
			_ = pr.output.Println(strconv.Itoa(current.File.Line))
			_ = pr.output.Println(current.File.LineContents)
		}
	}
	_ = pr.output.Println(current.Author.Login)
	_ = pr.output.Println(current.Body)
}

func (pr *PRAction) PrintState() {
	_ = pr.output.Println(pr.State.MergeStatus)
	_ = pr.output.Println(pr.State.ConflictStatus)
	for reviewState, names := range pr.State.Reviews {
		_ = pr.output.Println(fmt.Sprintf("%s %s", reviewState, strings.Join(names, " ")))
	}
	for _, status := range pr.State.Statuses {
		_ = pr.output.Println(fmt.Sprintf("Check %s %s", status.Name, status.Conclusion))
	}
}
