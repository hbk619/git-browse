package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/hbk619/git-browse/internal/git"
	"os/exec"
)

type Api interface {
	LoadGitHubAPIJSON(command string) ([]byte, error)
	RunCommand(command string) (string, error)
}

type GHApi struct{}

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
		errorMessage := ""
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
