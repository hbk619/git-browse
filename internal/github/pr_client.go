package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hbk619/git-browse/internal/git"
	"github.com/hbk619/git-browse/internal/github/graphql"
	"github.com/hbk619/git-browse/internal/requests"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

type PullRequestClient interface {
	GetPRDetails(repo *git.Repo, verbose bool) (*git.PR, error)
	GetRepoDetails() (*git.Repo, error)
	Reply(repo *git.Repo, contents string, comment *git.Comment) error
	Resolve(comment *git.Comment) error
}

type GetReviewCommentsQuery struct {
	PullRequestId int
	Owner         string
	RepoName      string
}

type PRClient struct {
	apiClient         Api
	commandLineClient requests.CommandLine
}

func NewPRClient(apiClient Api, line requests.CommandLine) *PRClient {
	return &PRClient{
		apiClient:         apiClient,
		commandLineClient: line,
	}
}

const MainThread = "main thread"

var mergeStatuses = map[string]string{
	"DIRTY":     "The merge commit cannot be cleanly created, try updating",
	"UNKNOWN":   "The state cannot currently be determined",
	"BLOCKED":   "The merge is blocked",
	"BEHIND":    "The branch is out of date",
	"UNSTABLE":  "Failing checks",
	"HAS_HOOKS": "Mergeable with passing checks and pre-receive hooks",
	"CLEAN":     "Mergeable and passing checks",
}

var mergeStates = map[string]string{
	"MERGEABLE":   "No conflicts",
	"CONFLICTING": "Merge conflicts",
	"UNKNOWN":     "The mergeability of the pull request is still being calculated",
}

func (gh *PRClient) GetPRDetails(repo *git.Repo, verbose bool) (*git.PR, error) {
	variables := map[string]interface{}{
		"PullRequestId": repo.PRNumber,
		"Owner":         repo.Owner,
		"RepoName":      repo.Name,
	}
	data, err := gh.apiClient.LoadGitHubGraphQLJSON(graphql.PRDetailsQuery(verbose), variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get pr details: %w", err)
	}

	var graphQLData git.GitHubData
	err = json.Unmarshal(data, &graphQLData)
	if err != nil {
		return nil, fmt.Errorf("failed to read pr details: %w", err)
	}

	prDetails := graphQLData.Data.Repository.PullRequest

	commentList := gh.createComments(&prDetails, verbose)

	return &git.PR{
		Comments: commentList,
		State:    gh.createState(verbose, prDetails),
		Title:    prDetails.Title,
	}, nil
}

func (gh *PRClient) createComments(response *git.PullRequest, verbose bool) []git.Comment {
	var commentList []git.Comment

	if response.Body != "" {
		commentList = append(commentList, git.Comment{
			Author: response.Author,
			Body:   response.Body,
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
			CreatedAt: response.CreatedAt,
		})
	}

	for _, comment := range response.Comments.Nodes {
		if comment.Body != "" {
			comment.File = git.File{
				FullPath: MainThread,
				FileName: MainThread,
			}
			commentList = append(commentList, comment)
		}
	}
	for _, comment := range response.Reviews.Nodes {
		if comment.Body != "" {
			comment.File = git.File{
				FullPath: MainThread,
				FileName: MainThread,
			}
			commentList = append(commentList, comment)
		}
	}
	gh.sortCommentsInPlace(commentList)
	reviewComments := gh.getThreadComments(response)
	commentList = append(commentList, reviewComments...)

	if verbose {
		commitComments := gh.getCommitDetails(&response.Commits)
		commentList = append(commentList, commitComments...)
	}
	return commentList
}

func (gh *PRClient) sortCommentsInPlace(commentList []git.Comment) {
	slices.SortFunc(commentList, func(i, j git.Comment) int {
		return time.Time.Compare(i.CreatedAt, j.CreatedAt)
	})
}

func (gh *PRClient) createState(verbose bool, prDetails git.PullRequest) git.State {
	state := git.State{}
	if verbose {
		reviewStatus := gh.getReviewStatuses(&prDetails.Reviews)
		state = git.State{
			Statuses:       prDetails.StatusCheckRollup.Contexts.Nodes,
			ConflictStatus: mergeStates[prDetails.Mergeable],
			MergeStatus:    mergeStatuses[prDetails.MergeStateStatus],
			Reviews:        reviewStatus,
		}
	}
	return state
}

func (gh *PRClient) getCommitDetails(commits *git.Commits) []git.Comment {
	var allComments []git.Comment
	for _, commitNode := range commits.Nodes {
		commit := commitNode.Commit
		for _, comment := range commit.Comments.Nodes {
			localComment := git.Comment{
				File: git.File{
					FullPath: commit.Oid,
					FileName: fmt.Sprintf("commit hash %s", commit.Oid),
				},
				Body:      comment.Body,
				Author:    comment.Author,
				CreatedAt: comment.CreatedAt,
			}
			allComments = append(allComments, localComment)
		}
	}
	return allComments
}

func (gh *PRClient) getReviewStatuses(response *git.Reviews) map[string][]string {
	reviewStatus := make(map[string][]string)
	alreadySeenReviewers := make(map[string]map[string]bool)
	for _, review := range response.Nodes {
		if alreadySeenReviewers[review.State] == nil {
			alreadySeenReviewers[review.State] = make(map[string]bool)
		}
		if !alreadySeenReviewers[review.State][review.Author.Login] {
			reviewStatus[review.State] = append(reviewStatus[review.State], review.Author.Login)
			alreadySeenReviewers[review.State][review.Author.Login] = true
		}
	}
	return reviewStatus
}

func (gh *PRClient) getThreadComments(graphQLData *git.PullRequest) []git.Comment {
	var comments []git.Comment
	threads := graphQLData.ReviewThreads.Nodes
	for _, thread := range threads {
		var threadComments []git.Comment
		for _, comment := range thread.Comments.Nodes {
			lineNumber := comment.File.Line
			if lineNumber == 0 {
				lineNumber = comment.OriginalLine
			}
			comment.File.FullPath = fmt.Sprintf("%s:%d", comment.File.Path, lineNumber)
			comment.File.FileName = filepath.Base(comment.File.Path)
			comment.File.Path = comment.File.Path[:len(comment.File.Path)-len(comment.File.FileName)]
			lines := strings.Split(comment.DiffHunk, "\n")
			comment.LineContents = lines[len(lines)-1]
			comment.Thread = git.Thread{ID: thread.ID, IsResolved: thread.IsResolved}

			threadComments = append(threadComments, comment)
		}
		gh.sortCommentsInPlace(threadComments)

		comments = append(comments, threadComments...)
	}
	return comments
}

var gitRepoRegex = regexp.MustCompile(`(?:git@|https://)[^:/]+[:/](?P<owner>[^/]+)/(?P<repo>.*)\.git`)

func (gh *PRClient) GetRepoDetails() (*git.Repo, error) {
	remote, err := gh.commandLineClient.Run("git config --get remote.origin.url")
	if err != nil || remote == "" {
		wd, _ := os.Getwd()
		return nil, fmt.Errorf(
			"not in a git repo, current directory: %s",
			wd,
		)
	}

	match := gitRepoRegex.FindStringSubmatch(remote)
	if match == nil {
		return nil, errors.New("could not parse git remote URL")
	}

	return &git.Repo{
		Owner: match[1],
		Name:  match[2],
	}, nil
}

func (gh *PRClient) Reply(repo *git.Repo, contents string, comment *git.Comment) error {
	if comment.Thread.ID == "" {
		command := fmt.Sprintf(`gh pr comment %d -b "%s"`, repo.PRNumber, contents)
		_, err := gh.commandLineClient.Run(command)
		return err

	}
	variables := map[string]interface{}{
		"threadId": comment.Thread.ID,
		"body":     contents,
	}
	_, err := gh.apiClient.LoadGitHubGraphQLJSON(graphql.AddCommentMutation, variables)
	return err
}

func (gh *PRClient) Resolve(comment *git.Comment) error {
	if comment.Thread.ID != "" {
		variables := map[string]interface{}{
			"threadId": comment.Thread.ID,
		}
		_, err := gh.apiClient.LoadGitHubGraphQLJSON(graphql.ResolveThreadMutation, variables)
		return err
	}
	return errors.New("cannot resolve a main or commit comment")
}
