package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/hbk619/git-browse/internal/git"
	"github.com/hbk619/git-browse/internal/github/graphql"
	mock_github "github.com/hbk619/git-browse/internal/github/mocks"
	mock_requests "github.com/hbk619/git-browse/internal/requests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	stdtime "time"
)

type PRServiceTestSuite struct {
	suite.Suite
	mockApi         *mock_github.MockApi
	ctrl            *gomock.Controller
	repo            *git.Repo
	prService       PRClient
	mockCommandLine *mock_requests.MockCommandLine
}

func (suite *PRServiceTestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockApi = mock_github.NewMockApi(suite.ctrl)
	suite.repo = &git.Repo{
		Owner:    "luigi",
		Name:     "castle",
		PRNumber: 123,
	}
	suite.mockCommandLine = mock_requests.NewMockCommandLine(suite.ctrl)
	suite.prService = PRClient{
		apiClient:         suite.mockApi,
		commandLineClient: suite.mockCommandLine,
	}
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_no_comments() {
	prDetails := git.PRDetails{
		Title:             "Test pr",
		Comments:          nil,
		Reviews:           nil,
		Body:              "",
		Author:            git.Author{Login: "Mario"},
		StatusCheckRollup: nil,
		Mergeable:         "",
		MergeStateStatus:  "",
	}
	marshalled, err := json.Marshal(prDetails)
	assert.NoError(suite.T(), err)

	expected := &git.PR{
		Comments: nil,
		State:    git.State{},
		Title:    "Test pr",
	}
	suite.mockCommandLine.EXPECT().
		Run("gh pr view 123 --json title,comments,reviews,body,author,createdAt").
		Return(string(marshalled), nil)
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}

	suite.mockApi.EXPECT().LoadGitHubGraphQLJSON(graphql.GetReviewCommentsQuery, variables).
		Return([]byte("{}"), nil)
	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_with_verbose() {
	prDetails := git.PRDetails{
		Title: "Test pr",
		Comments: []git.Comment{{
			Id: "awdasdadad",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Rraaawwww",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:38:47Z"),
		}, {
			Id: "lkmoimiom",
			Author: git.Author{
				Login: "Yoshi",
			},
			Body:      "Yum!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T22:38:47Z"),
		}},
		Reviews: []git.Comment{{
			Id: "23213213",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Great start",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:58:47Z"),
			State:     "COMMENTED",
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Gone down hill!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:38:47Z"),
			State:     "APPROVED",
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Keep it up!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:48:47Z"),
			State:     "COMMENTED",
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Wonderful!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-24T22:48:47Z"),
			State:     "COMMENTED",
		}},
		Body:      "My wonderful work",
		Author:    git.Author{Login: "Mario"},
		CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-20T22:38:47Z"),
		StatusCheckRollup: []git.Status{{
			Name:       "Test",
			Conclusion: "SUCCESS",
		}, {
			Name:       "Test1",
			Conclusion: "FAILURE",
		}},
		Mergeable:        "CONFLICTING",
		MergeStateStatus: "BLOCKED",
	}
	marshalled, err := json.Marshal(prDetails)
	assert.NoError(suite.T(), err)

	expected := &git.PR{
		Comments: []git.Comment{{
			Author: git.Author{
				Login: "Mario",
			},
			Body:      "My wonderful work",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-20T22:38:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "awdasdadad",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Rraaawwww",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:38:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "23213213",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Great start",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:58:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
			State: "COMMENTED",
		}, {
			Id: "lkmoimiom",
			Author: git.Author{
				Login: "Yoshi",
			},
			Body:      "Yum!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T22:38:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Gone down hill!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:38:47Z"),
			State:     "APPROVED",
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Keep it up!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:48:47Z"),
			State:     "COMMENTED",
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Wonderful!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-24T22:48:47Z"),
			State:     "COMMENTED",
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}},
		State: git.State{
			Reviews:        map[string][]string{"APPROVED": {"Peach"}, "COMMENTED": {"Peach", "Bowser"}},
			MergeStatus:    "The merge is blocked",
			ConflictStatus: "Merge conflicts",
			Statuses: []git.Status{{
				Name:       "Test",
				Conclusion: "SUCCESS",
			}, {
				Name:       "Test1",
				Conclusion: "FAILURE",
			}},
		},
		Title: "Test pr",
	}
	suite.mockCommandLine.EXPECT().
		Run("gh pr view 123 --json title,comments,reviews,body,author,createdAt,mergeStateStatus,mergeable,state,statusCheckRollup").
		Return(string(marshalled), nil)
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}

	suite.mockApi.EXPECT().LoadGitHubGraphQLJSON(graphql.GetReviewCommentsQuery, variables).
		Return([]byte("{}"), nil)
	details, err := suite.prService.GetPRDetails(suite.repo, true)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_pr_not_found() {
	suite.mockCommandLine.EXPECT().
		Run("gh pr view 123 --json title,comments,reviews,body,author,createdAt").
		Return("", nil)

	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.ErrorContains(suite.T(), err, "pull request not found")
	assert.Nil(suite.T(), details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_invalid_json_error() {
	suite.mockCommandLine.EXPECT().
		Run("gh pr view 123 --json title,comments,reviews,body,author,createdAt").
		Return("adsa", nil)

	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.ErrorContains(suite.T(), err, "failed to create json from pr")
	assert.Nil(suite.T(), details)
}
func (suite *PRServiceTestSuite) TestPRService_getPrDetails_review_comments_error() {
	prDetails := git.PRDetails{
		Title: "Test pr",
		Comments: []git.Comment{{
			Id: "awdasdadad",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Rraaawwww",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:38:47Z"),
		}, {
			Id: "lkmoimiom",
			Author: git.Author{
				Login: "Yoshi",
			},
			Body:      "Yum!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T22:38:47Z"),
		}},
		Reviews: []git.Comment{{
			Id: "23213213",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Great start",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:58:47Z"),
			State:     "COMMENTED",
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Gone down hill!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:38:47Z"),
			State:     "COMMENTED",
		}},
		Body:              "",
		Author:            git.Author{Login: "Mario"},
		StatusCheckRollup: nil,
		Mergeable:         "",
		MergeStateStatus:  "",
	}
	marshalled, err := json.Marshal(prDetails)
	assert.NoError(suite.T(), err)

	expected := errors.New("ahhhhh")
	suite.mockCommandLine.EXPECT().
		Run("gh pr view 123 --json title,comments,reviews,body,author,createdAt").
		Return(string(marshalled), nil)
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}

	suite.mockApi.EXPECT().LoadGitHubGraphQLJSON(graphql.GetReviewCommentsQuery, variables).
		Return(nil, expected)
	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.ErrorContains(suite.T(), err, expected.Error())
	assert.Nil(suite.T(), details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_comments() {
	prDetails := git.PRDetails{
		Title: "Test pr",
		Comments: []git.Comment{{
			Id: "awdasdadad",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Rraaawwww",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:38:47Z"),
		}, {
			Id: "lkmoimiom",
			Author: git.Author{
				Login: "Yoshi",
			},
			Body:      "Yum!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T22:38:47Z"),
		}},
		Reviews: []git.Comment{{
			Id: "23213213",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Great start",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:58:47Z"),
			State:     "COMMENTED",
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Gone down hill!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:38:47Z"),
			State:     "COMMENTED",
		}},
		Body:              "",
		Author:            git.Author{Login: "Mario"},
		StatusCheckRollup: nil,
		Mergeable:         "",
		MergeStateStatus:  "",
	}
	marshalled, err := json.Marshal(prDetails)
	assert.NoError(suite.T(), err)

	expected := &git.PR{
		Comments: []git.Comment{{
			Id: "awdasdadad",
			Author: git.Author{
				Login: "Bowser",
			},
			Body:      "Rraaawwww",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:38:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "23213213",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Great start",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T21:58:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
			State: "COMMENTED",
		}, {
			Id: "lkmoimiom",
			Author: git.Author{
				Login: "Yoshi",
			},
			Body:      "Yum!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-22T22:38:47Z"),
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "67867867867",
			Author: git.Author{
				Login: "Peach",
			},
			Body:      "Gone down hill!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2025-02-23T22:38:47Z"),
			State:     "COMMENTED",
			File: git.File{
				FullPath: MainThread,
				FileName: MainThread,
			},
		}, {
			Id: "ABCD_kDOOLvWJM5lOKjk",
			Author: git.Author{
				Login: "wario",
			},
			Body:      "Looking good!",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2024-07-31T09:34:11Z"),
			File: git.File{
				FullPath:     ".github/workflows/ci.yaml:2",
				Path:         ".github/workflows/",
				FileName:     "ci.yaml",
				OriginalLine: 2,
				DiffHunk:     "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]",
				LineContents: "+on: [push]",
				Line:         2,
			},
			Thread: git.Thread{
				IsResolved: false,
				ID:         "ABCD_kdER4tvWJM5AsoQq",
			},
		}, {
			Id: "PDDD_kwDOKtvW309DSOXr_",
			Author: git.Author{
				Login: "wario",
			},
			Body:      "this is a line comment not in a review",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2024-07-31T10:15:10Z"),
			File: git.File{
				FullPath:     ".github/workflows/ci.yaml:6",
				Path:         ".github/workflows/",
				FileName:     "ci.yaml",
				OriginalLine: 6,
				DiffHunk:     "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]\n+jobs:\n+  check-bats-version:\n+    runs-on: ubuntu-latest\n+    steps:",
				LineContents: "+    steps:",
				Line:         6,
			},
			Thread: git.Thread{
				IsResolved: true,
				ID:         "ABCD_kwDOKtvOWM5Aswyn",
			},
		}, {
			Id: "PRRD_kwDOKtvW3900DSOXr_",
			Author: git.Author{
				Login: "mario",
			},
			Body:      "this is a reply",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2024-07-31T11:15:10Z"),
			File: git.File{
				FullPath:     ".github/workflows/ci.yaml:6",
				Path:         ".github/workflows/",
				FileName:     "ci.yaml",
				OriginalLine: 6,
				DiffHunk:     "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]\n+jobs:\n+  check-bats-version:\n+    runs-on: ubuntu-latest\n+    steps:",
				LineContents: "+    steps:",
				Line:         6,
			},
			Thread: git.Thread{
				IsResolved: true,
				ID:         "ABCD_kwDOKtvOWM5Aswyn",
			},
		}},
		State: git.State{},
		Title: "Test pr",
	}
	suite.mockCommandLine.EXPECT().
		Run("gh pr view 123 --json title,comments,reviews,body,author,createdAt").
		Return(string(marshalled), nil)
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}

	suite.mockApi.EXPECT().LoadGitHubGraphQLJSON(graphql.GetReviewCommentsQuery, variables).
		Return([]byte(`{
  "data": {
    "repository": {
      "pullRequest": {
        "reviewThreads": {
          "nodes": [
            {
              "id": "ABCD_kdER4tvWJM5AsoQq",
              "isResolved": false,
              "comments": {
                "nodes": [
                  {
                    "id": "ABCD_kDOOLvWJM5lOKjk",
                    "body": "Looking good!",
                    "author": {
                      "login": "wario"
                    },
                    "originalLine": 2,
                    "originalStartLine": null,
                    "path": ".github/workflows/ci.yaml",
                    "line": 2,
                    "diffHunk": "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]",
                    "outdated": false,
                    "createdAt": "2024-07-31T09:34:11Z"
                  }
                ],
                "pageInfo": {
                  "hasNextPage": false
                }
              }
            },
            {
              "id": "ABCD_kwDOKtvOWM5Aswyn",
              "isResolved": true,
              "comments": {
                "nodes": [
                  {
                    "id": "PRRD_kwDOKtvW3900DSOXr_",
                    "body": "this is a reply",
                    "author": {
                      "login": "mario"
                    },
                    "originalLine": 6,
                    "originalStartLine": null,
                    "path": ".github/workflows/ci.yaml",
                    "line": 6,
                    "diffHunk": "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]\n+jobs:\n+  check-bats-version:\n+    runs-on: ubuntu-latest\n+    steps:",
                    "outdated": false,
                    "createdAt": "2024-07-31T11:15:10Z"
                  },
                  {
                    "id": "PDDD_kwDOKtvW309DSOXr_",
                    "body": "this is a line comment not in a review",
                    "author": {
                      "login": "wario"
                    },
                    "originalLine": 6,
                    "originalStartLine": null,
                    "path": ".github/workflows/ci.yaml",
                    "line": 6,
                    "diffHunk": "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]\n+jobs:\n+  check-bats-version:\n+    runs-on: ubuntu-latest\n+    steps:",
                    "outdated": false,
                    "createdAt": "2024-07-31T10:15:10Z"
                  }
                ],
                "pageInfo": {
                  "hasNextPage": false
                }
              }
            }
          ]
        }
      }
    }
  }
}`), nil)
	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestGetRepoDetails_ValidSSHURL() {
	suite.mockCommandLine.EXPECT().
		Run("git config --get remote.origin.url").
		Return("git@github.com:peach/repo-2.git", nil)

	repo, err := suite.prService.GetRepoDetails()

	suite.NoError(err)
	suite.Equal(&git.Repo{
		Owner: "peach",
		Name:  "repo-2",
	}, repo)
}

func (suite *PRServiceTestSuite) TestGetRepoDetails_ValidHTTPSURL() {
	suite.mockCommandLine.EXPECT().
		Run("git config --get remote.origin.url").
		Return("https://git.com/mario/castle.git", nil)

	repo, err := suite.prService.GetRepoDetails()

	suite.NoError(err)
	suite.Equal(&git.Repo{
		Owner: "mario",
		Name:  "castle",
	}, repo)
}

func (suite *PRServiceTestSuite) TestGetRepoDetails_CommandError() {
	suite.mockCommandLine.EXPECT().
		Run("git config --get remote.origin.url").
		Return("", errors.New("command failed"))

	repo, err := suite.prService.GetRepoDetails()

	suite.ErrorContains(err, "not in a git repo, current directory:")
	suite.Nil(repo)
}

func (suite *PRServiceTestSuite) TestGetRepoDetails_EmptyRemoteURL() {
	suite.mockCommandLine.EXPECT().
		Run("git config --get remote.origin.url").
		Return("", nil)

	repo, err := suite.prService.GetRepoDetails()

	suite.ErrorContains(err, "not in a git repo, current directory:")
	suite.Nil(repo)
}

func (suite *PRServiceTestSuite) TestGetRepoDetails_InvalidRemoteURL() {
	suite.mockCommandLine.EXPECT().
		Run("git config --get remote.origin.url").
		Return("invalid-url", nil)

	repo, err := suite.prService.GetRepoDetails()

	suite.ErrorContains(err, "could not parse git remote URL")
	suite.Nil(repo)
}

func (suite *PRServiceTestSuite) TestGetCommitComments_ValidInput() {
	getCommitsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/pulls/123/commits"`
	getCommentsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/commits/%s/comments"`

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(getCommitsCommand).
		Return([]byte(`[[{"sha": "sha123"},{"sha": "sha456"},{"sha": "sha7789"}]]`), nil)

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(fmt.Sprintf(getCommentsCommand, "sha123")).
		Return([]byte(`[[{"body": "Great work!", "created_at": "2025-04-12T10:00:00Z"}]]`), nil)
	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(fmt.Sprintf(getCommentsCommand, "sha456")).
		Return([]byte("[[]]"), nil)
	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(fmt.Sprintf(getCommentsCommand, "sha7789")).
		Return([]byte(`[[{"body": "I've seen better!", "created_at": "2025-04-12T19:00:00Z"}]]`), nil)

	comments, err := suite.prService.GetCommitComments("luigi", "mansion", 123)

	expectedComment1 := git.Comment{
		Body:      "Great work!",
		Created:   timeMustParse(stdtime.RFC3339, "2025-04-12T10:00:00Z"),
		CreatedAt: timeMustParse(stdtime.RFC3339, "2025-04-12T10:00:00Z"),
		File: git.File{
			FullPath: "sha123",
			FileName: "commit hash sha123",
		},
	}
	expectedComment2 := git.Comment{
		Body:      "I've seen better!",
		Created:   timeMustParse(stdtime.RFC3339, "2025-04-12T19:00:00Z"),
		CreatedAt: timeMustParse(stdtime.RFC3339, "2025-04-12T19:00:00Z"),
		File: git.File{
			FullPath: "sha7789",
			FileName: "commit hash sha7789",
		},
	}
	suite.NoError(err)
	suite.Len(comments, 2)
	suite.Equal(expectedComment1, comments[0])
	suite.Equal(expectedComment2, comments[1])
}

func (suite *PRServiceTestSuite) TestGetCommitComments_CommitsNotFound() {
	getCommitsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/pulls/123/commits"`

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(getCommitsCommand).
		Return(nil, errors.New("pull request not found"))

	comments, err := suite.prService.GetCommitComments("luigi", "mansion", 123)

	suite.Error(err)
	suite.Contains(err.Error(), "pull request not found")
	suite.Nil(comments)
}

func (suite *PRServiceTestSuite) TestGetCommitComments_NoCommentsForCommit() {
	getCommitsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/pulls/123/commits"`
	getCommentsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/commits/sha123/comments"`

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(getCommitsCommand).
		Return([]byte(`[[{"sha": "sha123"}]]`), nil)

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(getCommentsCommand).
		Return([]byte("[[]]"), nil)

	comments, err := suite.prService.GetCommitComments("luigi", "mansion", 123)

	suite.NoError(err)
	suite.Empty(comments)
}

func (suite *PRServiceTestSuite) TestGetCommitComments_ErrorInFetchingComments() {
	getCommitsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/pulls/123/commits"`
	getCommentsCommand := `gh api --paginate --slurp  -H "X-GitHub-Api-Version: 2022-11-28" \
-H "Accept: application/vnd.github+json" "/repos/luigi/mansion/commits/sha123/comments"`

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(getCommitsCommand).
		Return([]byte(`[[{"sha": "sha123"}]]`), nil)

	suite.mockApi.EXPECT().
		LoadGitHubAPIJSON(getCommentsCommand).
		Return(nil, errors.New("commit not found"))

	comments, err := suite.prService.GetCommitComments("luigi", "mansion", 123)

	suite.Error(err)
	suite.Contains(err.Error(), "commit not found")
	suite.Nil(comments)
}

func (suite *PRServiceTestSuite) TestReply_main_thread() {
	suite.mockCommandLine.EXPECT().
		Run(`gh pr comment 2 -b "Thank you"`).
		Return("", nil)
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	err := suite.prService.Reply(repo, "Thank you", &git.Comment{Id: "PDDD_e43oidmdm"})
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestReply_thread() {
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
		"body":     "Thank you",
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.AddCommentMutation, gomock.Eq(variables)).
		Return([]byte{}, nil)
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	err := suite.prService.Reply(repo, "Thank you", &git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"})
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestReply_has_error() {
	expected := errors.New("error")
	suite.mockCommandLine.EXPECT().
		Run(`gh pr comment 2 -b "Thank you"`).
		Return("", expected)
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	err := suite.prService.Reply(repo, "Thank you", &git.Comment{Id: "PDDD_e43oidmdm"})
	suite.ErrorIs(expected, err)
}

func (suite *PRServiceTestSuite) TestResolve_main_thread() {
	err := suite.prService.Resolve(&git.Comment{Id: "PDDD_e43oidmdm"})
	suite.ErrorContains(err, "cannot resolve a main or commit comment")
}

func (suite *PRServiceTestSuite) TestResolve_thread() {
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.ResolveThreadMutation, gomock.Eq(variables)).
		Return([]byte{}, nil)
	err := suite.prService.Resolve(&git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"})
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestResolve_has_error() {
	expected := errors.New("error")
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.ResolveThreadMutation, gomock.Eq(variables)).
		Return([]byte{}, expected)
	err := suite.prService.Resolve(&git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"})
	suite.ErrorIs(expected, err)
}

func TestPRServiceSuite(t *testing.T) {
	suite.Run(t, new(PRServiceTestSuite))
}

func timeMustParse(layout string, str string) time.Time {
	parse, err := time.Parse(layout, str)
	if err != nil {
		panic(err)
	}
	return parse
}
