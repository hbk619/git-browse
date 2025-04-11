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
	}

	File struct {
		FullPath     string
		Path         string
		Line         int
		LineContents string
		FileName     string
	}

	Comment struct {
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
		ThreadId       string
		FileDetails    File
	}
	Author struct {
		Login string
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
