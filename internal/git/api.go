package git

import (
	"encoding/json"
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
		Sha string
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
		Created        time.Time `json:"created_at"`
		SubmittedAt    time.Time
		Author         Author
		User           Author
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

	PullRequest struct {
		ReviewThreads ReviewThreads
	}

	GitHubData struct {
		Data struct {
			Repository Repository
		}
	}

	PRDetails struct {
		Title             string
		Comments          []Comment
		Reviews           []Comment
		Body              string
		Author            Author
		CreatedAt         time.Time
		StatusCheckRollup []Status
		Mergeable         string
		MergeStateStatus  string
	}
)

func (comment *Comment) UnmarshalJSON(data []byte) error {
	type C Comment
	if err := json.Unmarshal(data, (*C)(comment)); err != nil {
		return err
	}

	if (comment.User != Author{}) {
		comment.Author = comment.User
	}
	if comment.CreatedAt.IsZero() {
		date := comment.SubmittedAt
		if date.IsZero() {
			date = comment.Created
		}
		comment.CreatedAt = date
	}
	return nil
}
