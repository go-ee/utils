package enum

type Literal interface {
	Name() string
	Ordinal() int
}