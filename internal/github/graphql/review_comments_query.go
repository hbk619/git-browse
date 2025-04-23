package graphql

var GetReviewCommentsQuery = `
query PullRequestComments($PullRequestId: Int!, $Owner: String!,$RepoName: String!) {
  repository(owner: $Owner, name: $RepoName) {
    pullRequest(number: $PullRequestId) {
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
    }
  }
}
`
