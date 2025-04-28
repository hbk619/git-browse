package graphql

var ResolveThreadMutation = `mutation ResolveReviewThread($threadId: ID!) {
  resolveReviewThread(input: {threadId: $threadId}) {
    thread {
      id
    }
  }
}`
