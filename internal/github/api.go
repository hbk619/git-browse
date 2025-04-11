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

func NewGHApi() *GHApi {
	return &GHApi{}
}

func (ghApi *GHApi) RunCommand(command string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (ghApi *GHApi) LoadGitHubAPIJSON(command string) ([]byte, error) {
	output, err := ghApi.RunCommand(command)
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
