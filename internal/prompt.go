package internal

import (
	"bufio"
	"io"
	"strings"

	"github.com/hbk619/gh-peruse/internal/filesystem"
)

type Prompt interface {
	String(label string) string
}

type Prompter struct {
	input  io.Reader
	output filesystem.Output
}

func NewPrompt(input io.Reader, output filesystem.Output) *Prompter {
	return &Prompter{
		input:  input,
		output: output,
	}
}

func (prompt *Prompter) String(label string) string {
	var s string
	r := bufio.NewReader(prompt.input)
	for {
		_ = prompt.output.Print(label + ": ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}
