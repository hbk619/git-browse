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
	GetCommitComments(repoOwner, repoName string, pullNumber int) ([]git.Comment, error)
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
	response, err := gh.getMainPrDetails(repo, verbose)
	if err != nil {
		return nil, err
	}

	commentList := gh.createComments(response)

	state := git.State{}
	if verbose {
		reviewStatus := gh.getReviewStatuses(response)
		state = git.State{
			Statuses:       response.StatusCheckRollup,
			ConflictStatus: mergeStates[response.Mergeable],
			MergeStatus:    mergeStatuses[response.MergeStateStatus],
			Reviews:        reviewStatus,
		}
	}

	reviewComments, err := gh.getReviewComments(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get review comments from pr %w", err)
	}

	commentList = append(commentList, reviewComments...)
	return &git.PR{
		Comments: commentList,
		State:    state,
		Title:    response.Title,
	}, nil
}

func (gh *PRClient) createComments(response *git.PRDetails) []git.Comment {
	apiComments := append(response.Comments, response.Reviews...)
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

	for _, comment := range apiComments {
		if comment.Body != "" {
			comment.File = git.File{
				FullPath: MainThread,
				FileName: MainThread,
			}
			commentList = append(commentList, comment)
		}
	}
	gh.sortCommentsInPlace(commentList)
	return commentList
}

func (gh *PRClient) sortCommentsInPlace(commentList []git.Comment) {
	slices.SortFunc(commentList, func(i, j git.Comment) int {
		return time.Time.Compare(i.CreatedAt, j.CreatedAt)
	})
}

func (gh *PRClient) getMainPrDetails(repo *git.Repo, verbose bool) (*git.PRDetails, error) {
	verboseFields := ""
	if verbose {
		verboseFields = ",mergeStateStatus,mergeable,state,statusCheckRollup"
	}

	getCommentsCommand := fmt.Sprintf("gh pr view %d --json title,comments,reviews,body,author,createdAt%s", repo.PRNumber, verboseFields)
	comments, err := gh.commandLineClient.Run(getCommentsCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to get pr details %w", err)
	}

	if len(comments) == 0 {
		return nil, errors.New("pull request not found")
	}

	var response git.PRDetails
	err = json.Unmarshal([]byte(comments), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create json from pr. Received %s %w", comments, err)
	}
	return &response, nil
}

func (gh *PRClient) getReviewStatuses(response *git.PRDetails) map[string][]string {
	reviewStatus := make(map[string][]string)
	alreadySeenReviewers := make(map[string]map[string]bool)
	for _, review := range response.Reviews {
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

func (gh *PRClient) getReviewComments(repo *git.Repo) ([]git.Comment, error) {
	variables := map[string]interface{}{
		"PullRequestId": repo.PRNumber,
		"Owner":         repo.Owner,
		"RepoName":      repo.Name,
	}

	data, err := gh.apiClient.LoadGitHubGraphQLJSON(graphql.GetReviewCommentsQuery, variables)

	if err != nil {
		return nil, fmt.Errorf("failed to load review comments %w", err)
	}

	var graphQLData git.GitHubData
	err = json.Unmarshal(data, &graphQLData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse review comments because: %w", err)
	}
	comments := gh.getThreadComments(graphQLData)
	return comments, nil
}

func (gh *PRClient) getThreadComments(graphQLData git.GitHubData) []git.Comment {
	var comments []git.Comment
	threads := graphQLData.Data.Repository.PullRequest.ReviewThreads.Nodes
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

func (gh *PRClient) GetCommitComments(repoOwner, repoName string, pullNumber int) ([]git.Comment, error) {
	getCommitsCommand := fmt.Sprintf(`gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/%s/%s/pulls/%d/commits"`, repoOwner, repoName, pullNumber)

	results, err := gh.apiClient.LoadGitHubAPIJSON(getCommitsCommand)
	if err != nil {
		return nil, err
	}

	var commits [][]git.Commit
	err = json.Unmarshal(results, &commits)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commits %w", err)
	}
	var allComments []git.Comment
	for _, result := range Flatten(commits) {
		commitSHA := result.Sha
		getCommentsCommand := fmt.Sprintf(`gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/%s/%s/commits/%s/comments"`, repoOwner, repoName, commitSHA)

		commitComments, err := gh.apiClient.LoadGitHubAPIJSON(getCommentsCommand)
		if err != nil {
			return nil, err
		}
		var comments [][]git.Comment
		err = json.Unmarshal(commitComments, &comments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse comments %w", err)
		}
		for _, comment := range Flatten(comments) {
			comment.File = git.File{
				FullPath: commitSHA,
				FileName: fmt.Sprintf("commit hash %s", commitSHA),
			}
			allComments = append(allComments, comment)
		}
	}

	slices.SortFunc(allComments, func(a, b git.Comment) int {
		return time.Time.Compare(a.CreatedAt, b.CreatedAt)
	})

	return allComments, nil
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

func Flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}
