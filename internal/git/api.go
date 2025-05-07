package git

import (
	"time"
)

type (
	Error struct {
		Message string
	}
	State struct {
		Statuses       []Status
		ConflictStatus string
		MergeStatus    string
		Reviews        map[string][]string
	}

	Commit struct {
		Comments Comments
		Oid      string
	}

	CommitNode struct {
		Commit Commit
	}

	Repo struct {
		Owner    string
		Name     string
		PRNumber int
	}

	PR struct {
		Comments []Comment
		State    State
		Title    string
		Id       string
	}

	Status struct {
		Name       string
		Conclusion string
	}
	Thread struct {
		IsResolved bool
		ID         string
	}

	File struct {
		FullPath     string
		Path         string
		Line         int
		LineContents string
		FileName     string
		OriginalLine int
		DiffHunk     string
	}

	Comment struct {
		File
		Body           string
		CreatedAt      time.Time
		Author         Author
		State          string
		Outdated       bool
		Thread         Thread
		Id             interface{}
		MergeStatus    string
		ConflictStatus string
		Reviews        []string
		Statuses       []Status
	}
	Author struct {
		Login string
	}
	ReviewThreads struct {
		Nodes []ThreadNode
	}

	ThreadNode struct {
		ID         string
		IsResolved bool
		Comments   Comments
	}

	Comments struct {
		Nodes []Comment
	}

	Repository struct {
		PullRequest PullRequest
	}

	Contexts struct {
		Nodes []Status
	}

	StatusCheckRollup struct {
		State    string
		Contexts Contexts
	}
	Reviews struct {
		Nodes []Comment
	}
	Review struct {
		Author Author
		State  string
	}

	Commits struct {
		Nodes []CommitNode
	}
	PullRequest struct {
		ReviewThreads     ReviewThreads
		Title             string
		Body              string
		Author            Author
		CreatedAt         time.Time
		StatusCheckRollup StatusCheckRollup
		Mergeable         string
		MergeStateStatus  string
		Comments          Comments
		Reviews           Reviews
		Commits           Commits
		Id                string
	}

	GitHubData struct {
		Repository Repository
	}
)
