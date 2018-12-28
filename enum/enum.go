package enum

import (
	"strings"
)

type Literal interface {
	Name() string
	Ordinal() int
}

type EnumBaseJson struct {
	Name string `json:"name"`
}

func Parse(name string, literals []Literal) (ret Literal, ok bool) {
	for _, lit := range literals {
		if strings.EqualFold(lit.Name(), name) {
			return lit, true
		}
	}
	return nil, true
}
