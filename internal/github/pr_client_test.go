package github

import (
	"errors"
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
	prDetails := `{
  "data": {
    "repository": {
      "pullRequest": {
        "commits": {
          "nodes": [
            {
              "commit": {
                "oid": "712000baad1a5b641d0308d45be98444b521333",
                "comments": {
                  "nodes": []
                }
              }
            },
            {
              "commit": {
                "oid": "b4dsd44214789e2ae1f9c83fb6cfeaea24fr424",
                "comments": null
              }
            },
            {
              "commit": {
                "oid": "9c637cc6b5d6099b443980daf91248c354321ss221",
                "comments": {
                  "nodes": []
                }
              }
            }
          ]
        },
        "reviews": null,
        "body": "",
        "author": {
          "login": "Mario"
        },
        "title": "Test pr",
        "createdAt": "2025-02-20T22:38:47Z",
        "reviewThreads": null,
        "comments": null
      }
    }
  }
}`

	expected := &git.PR{
		Comments: nil,
		State:    git.State{},
		Title:    "Test pr",
	}

	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.PRDetailsQuery(false), variables).
		Return([]byte(prDetails), nil)
	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_with_verbose() {
	prDetails := `{
  "data": {
    "repository": {
      "pullRequest": {
        "commits": {
          "nodes": [
            {
              "commit": {
                "oid": "712000baad1ee2b6385d0308d45be98444b521333",
                "comments": {
                  "nodes": []
                }
              }
            },
            {
              "commit": {
                "oid": "b45facb711478f4eae1f9c83fb6cfeaea24fr224",
                "comments": {
                  "nodes": [
                    {
                      "body": "This is a commit comment",
                      "author": {
                        "login": "Mario"
                      },
                      "createdAt": "2024-07-23T09:30:30Z"
                    }
                  ]
                }
              }
            },
            {
              "commit": {
                "oid": "9c637cc6b5d6099b76980daf91248c35090j1es211",
                "comments": {
                  "nodes": []
                }
              }
            }
          ]
        },
        "reviews": {
          "pageInfo": {
            "hasNextPage": false
          },
          "nodes": [
            {
              "createdAt": "2025-02-22T21:58:47Z",
              "body": "Great start",
              "state": "COMMENTED",
              "author": {
                "login": "Peach"
              }
			},
            {
			  "author": {
                "login": "Peach"
              },
              "state": "APPROVED",
              "body": "Gone down hill!",
              "createdAt": "2025-02-23T22:38:47Z"
            },
            {
              "author": {
                "login": "Bowser"
              },
              "state": "COMMENTED",
              "body": "Keep it up!",
              "createdAt": "2025-02-23T22:48:47Z"
            },
            {
              "author": {
                "login": "Bowser"
              },
              "state": "COMMENTED",
              "body": "Wonderful!",
              "createdAt": "2025-02-24T22:48:47Z"
            }
          ]
        },
        "statusCheckRollup": {
          "state": "FAILURE",
          "contexts": {
            "nodes": [
              {
                "status": "COMPLETED",
                "name": "Test",
                "conclusion": "SUCCESS"
              },
              {
                "status": "COMPLETED",
                "name": "Test1",
                "conclusion": "FAILURE"
              }
            ]
          }
        },
        "mergeable": "CONFLICTING",
        "mergeStateStatus": "BLOCKED",
        "body": "My wonderful work",
        "author": {
          "login": "Mario"
        },
        "title": "Test pr",
        "createdAt": "2025-02-20T22:38:47Z",
        "reviewThreads": null,
        "comments": {
          "pageInfo": {
            "hasNextPage": false
          },
          "nodes": [
            {
              "id": "awdasdadad",
              "createdAt": "2025-02-22T21:38:47Z",
              "body": "Rraaawwww",
              "author": {
                "login": "Bowser"
              }
			},
            {
              "id": "lkmoimiom",
              "createdAt": "2025-02-22T22:38:47Z",
              "body": "Yum!",
              "author": {
                "login": "Yoshi"
              }
            }
          ]
        }
      }
    }
  }
}`

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
		}, {
			Author: git.Author{
				Login: "Mario",
			},
			Body:      "This is a commit comment",
			CreatedAt: timeMustParse(stdtime.RFC3339, "2024-07-23T09:30:30Z"),
			File: git.File{
				FullPath: "b45facb711478f4eae1f9c83fb6cfeaea24fr224",
				FileName: "commit hash b45facb711478f4eae1f9c83fb6cfeaea24fr224",
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
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.PRDetailsQuery(true), variables).
		Return([]byte(prDetails), nil)

	details, err := suite.prService.GetPRDetails(suite.repo, true)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_pr_not_found() {
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}
	expected := errors.New("failed to find pr")
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.PRDetailsQuery(false), variables).
		Return(nil, expected)

	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.ErrorContains(suite.T(), err, expected.Error())
	assert.Nil(suite.T(), details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_comments_with_threads() {
	prDetails := `{
  "data": {
    "repository": {
      "pullRequest": {

        "reviews": {
          "pageInfo": {
            "hasNextPage": false
          },
          "nodes": [
            {
              "createdAt": "2025-02-22T21:58:47Z",
              "body": "Great start",
              "state": "COMMENTED",
              "author": {
                "login": "Peach"
              }
			},
            {
			  "author": {
                "login": "Peach"
              },
              "state": "COMMENTED",
              "body": "Gone down hill!",
              "createdAt": "2025-02-23T22:38:47Z"
            }
          ]
        },
        "body": "",
        "author": {
          "login": "Mario"
        },
        "title": "Test pr",
        "createdAt": "2025-02-20T22:38:47Z",
        "reviewThreads": {
          "nodes" : [ {
            "id" : "ABCD_kdER4tvWJM5AsoQq",
            "isResolved" : false,
            "comments" : {
              "nodes" : [ {
                "id" : "ABCD_kDOOLvWJM5lOKjk",
                "body" : "Looking good!",
                "author" : {
                  "login" : "wario"
                },
                "originalLine" : 2,
                "originalStartLine" : null,
                "path" : ".github/workflows/ci.yaml",
                "line" : 2,
                "diffHunk" : "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]",
                "outdated" : false,
                "createdAt" : "2024-07-31T09:34:11Z"
              } ],
              "pageInfo" : {
                "hasNextPage" : false
              }
            }
          }, {
            "id" : "ABCD_kwDOKtvOWM5Aswyn",
            "isResolved" : true,
            "comments" : {
              "nodes" : [ 
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
				}, {
                "id" : "PDDD_kwDOKtvW309DSOXr_",
                "body" : "this is a line comment not in a review",
                "author" : {
                  "login" : "wario"
                },
                "originalLine" : 6,
                "originalStartLine" : null,
                "path" : ".github/workflows/ci.yaml",
                "line" : 6,
                "diffHunk" : "@@ -0,0 +1,8 @@\n+name: things\n+on: [push]\n+jobs:\n+  check-bats-version:\n+    runs-on: ubuntu-latest\n+    steps:",
                "outdated" : false,
                "createdAt" : "2024-07-31T10:15:10Z"
              } ],
              "pageInfo" : {
                "hasNextPage" : false
              }
            }
          }
		]
        },
        "comments": {
          "pageInfo": {
            "hasNextPage": false
          },
          "nodes": [
            {
              "id": "awdasdadad",
              "createdAt": "2025-02-22T21:38:47Z",
              "body": "Rraaawwww",
              "author": {
                "login": "Bowser"
              }
			},
            {
              "id": "lkmoimiom",
              "createdAt": "2025-02-22T22:38:47Z",
              "body": "Yum!",
              "author": {
                "login": "Yoshi"
              }
            }
          ]
        }
      }
    }
  }
}`

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
	variables := map[string]interface{}{
		"PullRequestId": suite.repo.PRNumber,
		"Owner":         suite.repo.Owner,
		"RepoName":      suite.repo.Name,
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.PRDetailsQuery(false), variables).
		Return([]byte(prDetails), nil)

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

func (suite *PRServiceTestSuite) TestReply_main_thread() {
	variables := map[string]interface{}{
		"pullRequestId": "asdsa2",
		"body":          "Thank you",
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.AddPRCommentMutation, gomock.Eq(variables), suite.repo).
		Return([]byte{}, nil)

	err := suite.prService.Reply(suite.repo, "Thank you", &git.Comment{Id: "PDDD_e43oidmdm"}, "asdsa2")
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestReply_thread() {
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
		"body":     "Thank you",
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.AddThreadCommentMutation, gomock.Eq(variables), suite.repo).
		Return([]byte{}, nil)
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	err := suite.prService.Reply(repo, "Thank you", &git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"})

	err := suite.prService.Reply(suite.repo, "Thank you", &git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"}, "")
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestReply_has_error() {
	expected := errors.New("error")
	variables := map[string]interface{}{
		"pullRequestId": "asdsa2",
		"body":          "Thank you",
	}
	suite.mockApi.EXPECT().
		LoadGitHubGraphQLJSON(graphql.AddPRCommentMutation, gomock.Eq(variables), suite.repo).
		Return([]byte{}, expected)

	err := suite.prService.Reply(suite.repo, "Thank you", &git.Comment{Id: "PDDD_e43oidmdm"}, "asdsa2")
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
