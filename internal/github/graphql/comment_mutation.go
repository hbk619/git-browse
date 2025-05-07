package graphql

var AddThreadCommentMutation = `mutation AddComment($threadId: ID!, $body: String!) {
  addPullRequestReviewThreadReply(input: {
    body: $body,
    pullRequestReviewThreadId: $threadId
  }) {clientMutationId}
}`

var AddPRCommentMutation = `mutation AddComment($pullRequestId: ID!, $body: String!) {
  addPullRequestReview(input: {
    body: $body
    pullRequestId: $pullRequestId
  }) {
    clientMutationId
  }
}`
