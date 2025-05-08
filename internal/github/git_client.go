package github

import "context"

type GitClient interface {
	CurrentBranch(ctx context.Context) (string, error)
}
