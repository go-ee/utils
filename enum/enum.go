package enum

import (
	"strings"
)

type Literal interface {
	Name() string
	Ordinal() int
}