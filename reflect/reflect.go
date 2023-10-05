package reflect

type LabeledCreator func(label string) interface{}

type Types struct {
	Label string
	Types map[string]LabeledCreator
}

func (o *Types) Resolve(name string) LabeledCreator {
	return o.Types[name]
}

func (o *Types) Register(name string, creator LabeledCreator) {
	o.Types[name] = creator
}

type LabeledTypes struct {
	LabeledTypes      map[string]*Types
	CacheResolveFirst map[string]LabeledCreator
}

func (o *LabeledTypes) Register(label string) (ret *Types) {
	ret = &Types{
		Label: label,
		Types: make(map[string]LabeledCreator),
	}
	o.LabeledTypes[label] = ret
	return
}

func (o *LabeledTypes) Resolve(namespace, typ string) (ret LabeledCreator) {
	if namespaceTypes := o.LabeledTypes[namespace]; namespaceTypes != nil {
		ret = namespaceTypes.Resolve(typ)
	}
	return
}

func (o *LabeledTypes) ResolveFirst(typ string) (ret LabeledCreator) {
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

func (o *LabeledTypes) resolveFirst(typ string) (ret LabeledCreator) {
	for _, types := range o.LabeledTypes {
		ret = types.Resolve(typ)
		if ret != nil {
			break
		}
	}
	return
}
