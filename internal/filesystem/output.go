package filesystem

import "fmt"

type (
	Output interface {
		Print(text string)
	}

	StdOut struct{}
)

func NewStdOut() *StdOut {
	return &StdOut{}
}

func (stdOut *StdOut) Print(text string) {
	fmt.Println(text)
}
