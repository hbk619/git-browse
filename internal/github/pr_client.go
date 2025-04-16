package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hbk619/git-browse/internal/git"
	"os"
	"regexp"
	"slices"
	"time"
)

type PullRequestClient interface {
	GetCommitComments(repoOwner, repoName string, pullNumber int) ([]git.Comment, error)
	GetMainPRDetails(pullNumber int, verbose bool) (*git.PR, error)
	GetRepoDetails() (*git.Repo, error)
}

type PRClient struct {
	apiClient Api
}

func NewPRClient(apiClient Api) *PRClient {
	return &PRClient{
		apiClient: apiClient,
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

func (gh *PRClient) GetMainPRDetails(pullNumber int, verbose bool) (*git.PR, error) {
	verboseFields := ""
	if verbose {
		verboseFields = ",mergeStateStatus,mergeable,state,statusCheckRollup"
	}

	getCommentsCommand := fmt.Sprintf("gh pr view %d --json title,comments,reviews,body,author,createdAt%s", pullNumber, verboseFields)
	comments, err := gh.apiClient.RunCommand(getCommentsCommand)
	if err != nil {
		return &git.PR{}, err
	}

	if len(comments) == 0 {
		return nil, errors.New("pull request not found")
	}

	var response git.PRDetails
	err = json.Unmarshal([]byte(comments), &response)
	if err != nil {
		return nil, err
	}

	apiComments := append(response.Comments, response.Reviews...)
	var commentList []git.Comment

	if response.Body != "" {
		commentList = append(commentList, git.Comment{
			Author: response.Author,
			Body:   response.Body,
			FileDetails: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
			CreatedAt: response.CreatedAt,
		})
	}

	for _, comment := range apiComments {
		if comment.Body != "" {
			comment.FileDetails = git.File{
				FullPath: MainThread,
				FileName: MainThread,
			}
			commentList = append(commentList, comment)
		}
	}

	slices.SortFunc(commentList, func(i, j git.Comment) int {
		return time.Time.Compare(i.CreatedAt, j.CreatedAt)
	})

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

	return &git.PR{
		Comments: commentList,
		State:    state,
		Title:    response.Title,
	}, nil
}

func (gh *PRClient) getReviewStatuses(response git.PRDetails) map[string][]string {
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

var gitRepoRegex = regexp.MustCompile(`(?:git@|https://)[^:/]+[:/](?P<owner>[^/]+)/(?P<repo>.*)\.git`)

func (gh *PRClient) GetRepoDetails() (*git.Repo, error) {
	remote, err := gh.apiClient.RunCommand("git config --get remote.origin.url")
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
			comment.FileDetails = git.File{
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

func Flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}
