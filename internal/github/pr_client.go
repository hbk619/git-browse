package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	githubql "github.com/cli/shurcooL-graphql"
	"github.com/hbk619/git-browse/internal/git"
	"github.com/hbk619/git-browse/internal/github/graphql"
	"github.com/hbk619/git-browse/internal/requests"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type PullRequestClient interface {
	DetectCurrentPR(repo *git.Repo) (int, error)
	GetPRDetails(repo *git.Repo, verbose bool) (*git.PR, error)
	GetRepoDetails() (repository.Repository, error)
	Resolve(comment *git.Comment) error
	Reply(contents string, comment *git.Comment, prId string) error
}

type GetReviewCommentsQuery struct {
	PullRequestId int
	Owner         string
	RepoName      string
}

type PRClient struct {
	graphQLClient requests.GraphQLClient
	gitClient     GitClient
}

func NewPRClient(apiClient requests.GraphQLClient, gitClient GitClient) *PRClient {
	return &PRClient{
		graphQLClient: apiClient,
		gitClient:     gitClient,
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

type GraphQLResponse struct {
	Data   interface{}
	Errors []api.GraphQLErrorItem
}

func (gh *PRClient) GetPRDetails(repo *git.Repo, verbose bool) (*git.PR, error) {
	variables := map[string]interface{}{
		"PullRequestId": githubql.Int(repo.PRNumber),
		"Owner":         githubql.String(repo.Owner),
		"RepoName":      githubql.String(repo.Name),
	}
	var response git.GitHubData
	err := gh.graphQLClient.Do(graphql.PRDetailsQuery(verbose), variables, &response)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch pr info %w", err)
	}

	prDetails := response.Repository.PullRequest

	commentList := gh.createComments(&prDetails, verbose)

	return &git.PR{
		Comments: commentList,
		State:    gh.createState(verbose, &prDetails),
		Title:    prDetails.Title,
		Id:       prDetails.Id,
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

func (gh *PRClient) createState(verbose bool, prDetails *git.PullRequest) git.State {
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

func (gh *PRClient) DetectCurrentPR(repo *git.Repo) (int, error) {
	branch, err := gh.gitClient.CurrentBranch(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get branch %w", err)
	}

	variables := map[string]interface{}{
		"BranchName": branch,
		"Owner":      repo.Owner,
		"RepoName":   repo.Name,
	}
	var prList git.GitHubData
	err = gh.graphQLClient.Do(graphql.GetPRForBranch, variables, &prList)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch pr %w", err)
	}

	if len(prList.Repository.PullRequests.Nodes) == 0 {
		return 0, fmt.Errorf("no pull request found for %s", branch)
	}

	if len(prList.Repository.PullRequests.Nodes) > 1 {
		return 0, fmt.Errorf("too many pull request found for %s", branch)
	}

	return prList.Repository.PullRequests.Nodes[0].Number, nil
}

func (gh *PRClient) GetRepoDetails() (repository.Repository, error) {
	return repository.Current()
}

func (gh *PRClient) Reply(contents string, comment *git.Comment, prId string) error {
	query := graphql.AddThreadCommentMutation
	variables := map[string]interface{}{
		"threadId": comment.Thread.ID,
		"body":     contents,
	}

	if comment.Thread.ID == "" {
		query = graphql.AddPRCommentMutation
		variables = map[string]interface{}{
			"body":          contents,
			"pullRequestId": prId,
		}
	}

	return gh.graphQLClient.Do(query, variables, nil)
}

func (gh *PRClient) Resolve(comment *git.Comment) error {
	if comment.Thread.ID != "" {
		variables := map[string]interface{}{
			"threadId": comment.Thread.ID,
		}

		var results GraphQLResponse
		return gh.graphQLClient.Do(graphql.ResolveThreadMutation, variables, results)
	}
	return errors.New("cannot resolve a main or commit comment")
}
