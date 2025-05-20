package graphql

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/git"
)

func PRDetailsQuery(verbose bool) string {
	verboseFields := ""
	if verbose {
		verboseFields = `commits(first: 100) {
			nodes {
			commit {
				oid 
				comments(first: 100) {
				nodes {
				  body
				  author {login}
				  createdAt
				}
			  }
			}
		  }
		}
	  statusCheckRollup {
		state
		contexts(first: 100) {
			nodes {
			... on CheckRun {
			  status
			  name
			  conclusion
			}
		  }

		}
	  }
	  mergeable
	  mergeStateStatus
`
	}
	return fmt.Sprintf(`
query PullRequestComments($PullRequestId: Int!, $Owner: String!,$RepoName: String!) {
  repository(owner: $Owner, name:$RepoName) {
    pullRequest(number: $PullRequestId) {
		%s
		id
      reviews(first: 100) {
			pageInfo {hasNextPage}
			nodes {
			  author {login}
			  state
			  body
			  createdAt
			}
		  }
      body
      author{login}
      title
      createdAt
      reviewThreads(first: 100) {
            nodes {
                id,
                isResolved,
                comments(first: 100) {
                    nodes {
                      id
                      body
                      author {
                        login
                      },
                      originalLine,
                      originalStartLine,
                      path,
                      line,
                      diffHunk,
                      outdated,
                      createdAt
                    }
                    pageInfo {
                      hasNextPage
                    }
              }
          }
      }
      comments(first: 100) {
        pageInfo { hasNextPage }
        nodes {
          id,
          createdAt,
          body,
          author {
            login
          }
        }
      }
    }
  }
}`, verboseFields)
}

var GetPRForBranch = `query GetPRForBranch($BranchName: String!, $Owner: String!, $RepoName: String!) {
  repository(owner:$Owner, name:$RepoName) {
    pullRequests(first: 10, headRefName:$BranchName) {
      nodes {
        number
      }
    }
  }
}`

func GetAllPRsFor(repo *git.Repo) string {
	repoUrl := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	return fmt.Sprintf(`query {
    search(
      type: ISSUE,
      query: "repo:%s author:@me state:open is:pr",
      first: 20
    ) {
      edges {
        node {
          ... on PullRequest {
          id
          number
          commits(first: 100) {
            nodes {
              commit {
                comments(first: 100) {
                  nodes {
                    body
                  }
                }
              }
            }
          }
          reviews(first: 100) {
            nodes {
                body
              }
          }
          reviewThreads(first: 100) {
            nodes {
              comments(first: 100) {
                nodes {
                  body
                }
              }
            }
          }
          comments(first: 100) {
            nodes {
              body
            }
          }
        }
      }
    }
  }
}`, repoUrl)
}
