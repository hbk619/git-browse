package new_comments

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/filesystem"
	"github.com/hbk619/gh-peruse/internal/git"
	"github.com/hbk619/gh-peruse/internal/github"
	"github.com/hbk619/gh-peruse/internal/history"
)

func CheckForNewComments(prClient github.PullRequestClient, historyService history.Storage, output filesystem.Output) error {
	repo, err := prClient.GetRepoDetails()
	if err != nil {
		return fmt.Errorf("failed to get repo info %w", err)
	}
	internalRepo := &git.Repo{
		Owner: repo.Owner,
		Name:  repo.Name,
	}
	prs, err := prClient.GetCommentCountForOwnedPRs(internalRepo)
	if err != nil {
		return fmt.Errorf("failed to get prs %w", err)
	}

	prHistory, err := historyService.Load()
	if err != nil {
		return fmt.Errorf("failed to load comments from history: %w", err)
	}

	var notifyError error
	for pr, count := range prs {
		oldCount := prHistory.Prs[pr].CommentCount
		if oldCount != count {
			msg := fmt.Sprintf("Pull request %d has new comments", pr)
			err = output.Println(msg)
			if err != nil {
				notifyError = err
			}

		}
	}
	return notifyError
}
