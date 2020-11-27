package enum

import (
	"strings"
)

type Literal interface {
	Name() string
	Ordinal() int
}

func Parse(name string, literals []Literal) (ret Literal, ok bool) {
	for _, lit := range literals {
		if strings.EqualFold(lit.Name(), name) {
			return lit, true
		}
	}
	return nil, false
}