package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hbk619/git-browse/internal/git"
	"github.com/hbk619/git-browse/internal/requests"
	"io"
	"net/http"
	"os"
	"strings"
)

type Api interface {
	LoadGitHubAPIJSON(command string) ([]byte, error)
	LoadGitHubGraphQLJSON(query string, variables map[string]interface{}) ([]byte, error)
}

type GraphQLError struct {
	Errors []git.Error
}

type GHApi struct {
	httpClient        requests.HTTPClient
	commandLineClient requests.CommandLine
}

func NewGHApi(client requests.HTTPClient, line requests.CommandLine) *GHApi {
	return &GHApi{
		httpClient:        client,
		commandLineClient: line,
	}
}

func (ghApi *GHApi) LoadGitHubAPIJSON(command string) ([]byte, error) {
	output, err := ghApi.commandLineClient.Run(command)
	if err != nil {
		return nil, err
	}

	var results []git.Error
	_ = json.Unmarshal([]byte(output), &results)

	if len(results) == 1 && results[0].Message != "" {
		errorMessage := results[0].Message
		switch results[0].Message {
		case "Not found":
			errorMessage = "pull request not found"
		case "No commit found":
			errorMessage = "commit not found"
		}
		return nil, errors.New(errorMessage)
	}

	return []byte(output), nil
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func (ghApi *GHApi) LoadGitHubGraphQLJSON(query string, variables map[string]interface{}) ([]byte, error) {
	graphqlURL := getEnv("GITHUB_GRAPHQL", "https://api.github.com/graphql")
	gqlRequest := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	body, err := json.Marshal(gqlRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create body for Github Graphql request because: %v", err)
	}
	req, err := http.NewRequest("POST", graphqlURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for Github Graphql because: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ghApi.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to get Github Graphql data because: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Github Graphql response because: %v", err)
	}

	var results GraphQLError
	_ = json.Unmarshal(data, &results)

	if len(results.Errors) > 0 {
		var errorMessage []string
		for _, e := range results.Errors {
			errorMessage = append(errorMessage, e.Message)
		}
		return nil, fmt.Errorf("failed to get Github graphql data %s", strings.Join(errorMessage, " "))
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errors.New(fmt.Sprintf("%s not found", graphqlURL))
	case http.StatusForbidden:
		return nil, errors.New(fmt.Sprintf("not allowed to view url %s", graphqlURL))
	case http.StatusUnauthorized:
		return nil, errors.New("bad auth token")
	default:
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("something bad happened %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
	}

	return data, nil
}
