package reflect

type LabeledCreator[T any, M any] func() (T, M)

type Types[T any, M any] struct {
	Label string
	Types map[string]LabeledCreator[T, M]
}

func (o *Types[T, M]) Resolve(name string) LabeledCreator[T, M] {
	return o.Types[name]
}

func (o *Types[T, M]) Register(name string, creator LabeledCreator[T, M]) {
	o.Types[name] = creator
}

func NewLabeledTypes[T any, M any]() *LabeledTypes[T, M] {
	return &LabeledTypes[T, M]{
		LabeledTypes:      make(map[string]*Types[T, M]),
		CacheResolveFirst: make(map[string]LabeledCreator[T, M]),
	}
}

type LabeledTypes[T any, M any] struct {
	LabeledTypes      map[string]*Types[T, M]
	CacheResolveFirst map[string]LabeledCreator[T, M]
}

func (o *LabeledTypes[T, M]) Register(label string) (ret *Types[T, M]) {
	ret = &Types[T, M]{
		Label: label,
		Types: make(map[string]LabeledCreator[T, M]),
	}
	o.LabeledTypes[label] = ret
	return
}

func (o *LabeledTypes[T, M]) Resolve(namespace, typ string) (ret LabeledCreator[T, M]) {
	if namespaceTypes := o.LabeledTypes[namespace]; namespaceTypes != nil {
		ret = namespaceTypes.Resolve(typ)
	}
	return
}

func (o *LabeledTypes[T, M]) ResolveFirst(typ string) (ret LabeledCreator[T, M]) {
	if o.CacheResolveFirst != nil {
		var ok bool
		if ret, ok = o.CacheResolveFirst[typ]; !ok {
			ret = o.resolveFirst(typ)
			o.CacheResolveFirst[typ] = ret
		}
	} else {
		ret = o.resolveFirst(typ)
	}
	return
}

func (o *LabeledTypes[T, M]) resolveFirst(typ string) (ret LabeledCreator[T, M]) {
	for _, types := range o.LabeledTypes {
		ret = types.Resolve(typ)
		if ret != nil {
			break
		}
	}
	return
}
