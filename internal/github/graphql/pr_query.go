package graphql

import "fmt"

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
