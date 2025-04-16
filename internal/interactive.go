package internal

import "fmt"

type Interactive struct {
	Index    int
	MaxIndex int
}

func (i *Interactive) Next(print func()) {
	if i.Index < i.MaxIndex {
		i.Index++
		print()
	} else {
		fmt.Println("Nothing here")
	}
}

func (i *Interactive) Previous(print func()) {
	if i.Index > 0 {
		i.Index--
		print()
	} else {
		fmt.Println("Nothing here")
	}
}

func (i *Interactive) Repeat(print func()) {
	print()
}
