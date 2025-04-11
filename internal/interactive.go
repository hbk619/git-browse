package internal

import "fmt"

type Interactive struct {
	Index    int
	MaxIndex int
}

func (i *Interactive) N(print func()) {
	if i.Index < i.MaxIndex {
		i.Index++
		print()
	} else {
		fmt.Println("Nothing here")
	}
}

func (i *Interactive) P(print func()) {
	if i.Index > 0 {
		i.Index--
		print()
	} else {
		fmt.Println("Nothing here")
	}
}

func (i *Interactive) R(print func()) {
	print()
}
