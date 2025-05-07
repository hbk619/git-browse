package requests

type (
	GraphQLClient interface {
		Do(query string, variables map[string]interface{}, response interface{}) error
	}
)
