package github

import (
	"encoding/json"
	"errors"
	githubql "github.com/cli/shurcooL-graphql"
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
	mockGraphQL   *mock_requests.MockGraphQLClient
	mockGitClient *mock_github.MockGitClient
	ctrl          *gomock.Controller
	repo          *git.Repo
	prService     PRClient
}

func (suite *PRServiceTestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockGraphQL = mock_requests.NewMockGraphQLClient(suite.ctrl)
	suite.mockGitClient = mock_github.NewMockGitClient(suite.ctrl)
	suite.repo = &git.Repo{
		Owner:    "luigi",
		Name:     "castle",
		PRNumber: 123,
	}
	suite.prService = PRClient{
		graphQLClient: suite.mockGraphQL,
		gitClient:     suite.mockGitClient,
	}
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_no_comments() {
	prDetails := `{
  "data": {
    "repository": {
      "pullRequest": {
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
		"PullRequestId": githubql.Int(suite.repo.PRNumber),
		"Owner":         githubql.String(suite.repo.Owner),
		"RepoName":      githubql.String(suite.repo.Name),
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.PRDetailsQuery(false), variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			gr := GraphQLResponse{Data: response}

			err := json.Unmarshal([]byte(prDetails), &gr)
			suite.NoError(err)
			return nil
		})
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
		"PullRequestId": githubql.Int(suite.repo.PRNumber),
		"Owner":         githubql.String(suite.repo.Owner),
		"RepoName":      githubql.String(suite.repo.Name),
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.PRDetailsQuery(true), variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			gr := GraphQLResponse{Data: response}
			err := json.Unmarshal([]byte(prDetails), &gr)
			suite.NoError(err)
			return nil
		})

	details, err := suite.prService.GetPRDetails(suite.repo, true)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestPRService_getPrDetails_error() {
	variables := map[string]interface{}{
		"PullRequestId": githubql.Int(suite.repo.PRNumber),
		"Owner":         githubql.String(suite.repo.Owner),
		"RepoName":      githubql.String(suite.repo.Name),
	}
	expected := errors.New("failed to find pr")
	suite.mockGraphQL.EXPECT().
		Do(graphql.PRDetailsQuery(false), variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			return expected
		})

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
		"PullRequestId": githubql.Int(suite.repo.PRNumber),
		"Owner":         githubql.String(suite.repo.Owner),
		"RepoName":      githubql.String(suite.repo.Name),
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.PRDetailsQuery(false), variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			gr := GraphQLResponse{Data: response}
			err := json.Unmarshal([]byte(prDetails), &gr)
			suite.NoError(err)
			return nil
		})

	details, err := suite.prService.GetPRDetails(suite.repo, false)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, details)
}

func (suite *PRServiceTestSuite) TestReply_main_thread() {
	variables := map[string]interface{}{
		"pullRequestId": "asdsa2",
		"body":          "Thank you",
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.AddPRCommentMutation, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			return nil
		})

	err := suite.prService.Reply("Thank you", &git.Comment{Id: "PDDD_e43oidmdm"}, "asdsa2")
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestReply_thread() {
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
		"body":     "Thank you",
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.AddThreadCommentMutation, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			return nil
		})

	err := suite.prService.Reply("Thank you", &git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"}, "")
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestReply_has_error() {
	expected := errors.New("error")
	variables := map[string]interface{}{
		"pullRequestId": "asdsa2",
		"body":          "Thank you",
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.AddPRCommentMutation, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			return expected
		})

	err := suite.prService.Reply("Thank you", &git.Comment{Id: "PDDD_e43oidmdm"}, "asdsa2")
	suite.ErrorIs(err, expected)
}

func (suite *PRServiceTestSuite) TestResolve_main_thread() {
	err := suite.prService.Resolve(&git.Comment{Id: "PDDD_e43oidmdm"})
	suite.ErrorContains(err, "cannot resolve a main or commit comment")
}

func (suite *PRServiceTestSuite) TestResolve_thread() {
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.ResolveThreadMutation, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			return nil
		})
	err := suite.prService.Resolve(&git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"})
	suite.NoError(err)
}

func (suite *PRServiceTestSuite) TestResolve_has_error() {
	expected := errors.New("error")
	variables := map[string]interface{}{
		"threadId": "P2323123dm",
	}
	suite.mockGraphQL.EXPECT().
		Do(graphql.ResolveThreadMutation, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			return expected
		})
	err := suite.prService.Resolve(&git.Comment{Thread: git.Thread{ID: "P2323123dm"}, Id: "PDDD_e43oidmdm"})
	suite.ErrorIs(err, expected)
}

func (suite *PRServiceTestSuite) TestDetectCurrentPR_has_error_getting_branch() {
	expected := errors.New("error")

	suite.mockGitClient.EXPECT().
		CurrentBranch(gomock.Any()).
		Return("", expected)
	prNumber, err := suite.prService.DetectCurrentPR(&git.Repo{
		Owner: "mario",
		Name:  "kart",
	})
	suite.ErrorIs(err, expected)
	suite.Equal(prNumber, 0)
}

func (suite *PRServiceTestSuite) TestDetectCurrentPR_has_error_getting_prs() {
	expected := errors.New("error")
	variables := map[string]interface{}{
		"BranchName": "branchy",
		"Owner":      "mario",
		"RepoName":   "kart",
	}
	suite.mockGitClient.EXPECT().
		CurrentBranch(gomock.Any()).
		Return("branchy", nil)
	suite.mockGraphQL.EXPECT().
		Do(graphql.GetPRForBranch, variables, gomock.Any()).
		Return(expected)
	prNumber, err := suite.prService.DetectCurrentPR(&git.Repo{
		Owner: "mario",
		Name:  "kart",
	})
	suite.ErrorIs(err, expected)
	suite.Equal(0, prNumber)
}

func (suite *PRServiceTestSuite) TestDetectCurrentPR_returns_pr_when_one_found() {
	variables := map[string]interface{}{
		"BranchName": "branchy",
		"Owner":      "mario",
		"RepoName":   "kart",
	}
	prDetails := `{
		"data": {
			"repository": {
				"pullRequests": {
					"nodes": [
						{
							"number": 2
						}
					]
				}
			}
		}
	}`
	suite.mockGitClient.EXPECT().
		CurrentBranch(gomock.Any()).
		Return("branchy", nil)
	suite.mockGraphQL.EXPECT().
		Do(graphql.GetPRForBranch, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			gr := GraphQLResponse{Data: response}

			err := json.Unmarshal([]byte(prDetails), &gr)
			suite.NoError(err)
			return nil
		})
	prNumber, err := suite.prService.DetectCurrentPR(&git.Repo{
		Owner: "mario",
		Name:  "kart",
	})
	suite.NoError(err)
	suite.Equal(2, prNumber)
}

func (suite *PRServiceTestSuite) TestDetectCurrentPR_returns_pr_error_when_more_than_one_found() {
	variables := map[string]interface{}{
		"BranchName": "branchy",
		"Owner":      "mario",
		"RepoName":   "kart",
	}
	prDetails := `{
		"data": {
			"repository": {
				"pullRequests": {
					"nodes": [
						{
							"number": 2
						},
						{
							"number": 4
						}
					]
				}
			}
		}
	}`
	suite.mockGitClient.EXPECT().
		CurrentBranch(gomock.Any()).
		Return("branchy", nil)
	suite.mockGraphQL.EXPECT().
		Do(graphql.GetPRForBranch, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			gr := GraphQLResponse{Data: response}

			err := json.Unmarshal([]byte(prDetails), &gr)
			suite.NoError(err)
			return nil
		})
	prNumber, err := suite.prService.DetectCurrentPR(&git.Repo{
		Owner: "mario",
		Name:  "kart",
	})
	suite.ErrorContains(err, "too many pull request found for branchy")
	suite.Equal(0, prNumber)
}

func (suite *PRServiceTestSuite) TestDetectCurrentPR_returns_pr_error_when_none_found() {
	variables := map[string]interface{}{
		"BranchName": "branchy",
		"Owner":      "mario",
		"RepoName":   "kart",
	}
	prDetails := `{
		"data": {
			"repository": {
				"pullRequests": {
					"nodes": []
				}
			}
		}
	}`
	suite.mockGitClient.EXPECT().
		CurrentBranch(gomock.Any()).
		Return("branchy", nil)
	suite.mockGraphQL.EXPECT().
		Do(graphql.GetPRForBranch, variables, gomock.Any()).
		DoAndReturn(func(query string, variables map[string]interface{}, response interface{}) error {
			gr := GraphQLResponse{Data: response}

			err := json.Unmarshal([]byte(prDetails), &gr)
			suite.NoError(err)
			return nil
		})
	prNumber, err := suite.prService.DetectCurrentPR(&git.Repo{
		Owner: "mario",
		Name:  "kart",
	})
	suite.ErrorContains(err, "no pull request found for branchy")
	suite.Equal(0, prNumber)
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
