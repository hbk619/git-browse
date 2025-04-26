package graphql

var AddCommentMutation = `mutation AddComment($threadId: ID!, $body: String!) {
  addPullRequestReviewThreadReply(input: {
    body: $body,
    pullRequestReviewThreadId: $threadId
  }) {clientMutationId}
}`
