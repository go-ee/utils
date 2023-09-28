package reflect

import "reflect"

type Types struct {
	Label string
	Types map[string]reflect.Type
}

func (o *Types) NewInstance(name string) (ret interface{}) {
	return reflect.New(o.Resolve(name)).Elem().Interface()
}

func (o *Types) Resolve(name string) reflect.Type {
	return o.Types[name]
}

func (o *Types) Register(name string, typedNil interface{}) {
	o.Types[name] = reflect.TypeOf(typedNil).Elem()
}

type LabeledTypes struct {
	LabeledTypes      map[string]*Types
	CacheResolveFirst map[string]reflect.Type
}

func (o *LabeledTypes) Register(label string) (ret *Types) {
	ret = &Types{
		Label: label,
		Types: make(map[string]reflect.Type),
	}
	o.LabeledTypes[label] = ret
	return
}

func (o *LabeledTypes) Resolve(namespace, typ string) (ret reflect.Type) {
	if namespaceTypes := o.LabeledTypes[namespace]; namespaceTypes != nil {
		ret = namespaceTypes.Resolve(typ)
	}
	return
}

func (o *LabeledTypes) ResolveFirst(typ string) (ret reflect.Type) {
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

func (o *LabeledTypes) resolveFirst(typ string) (ret reflect.Type) {
	for _, types := range o.LabeledTypes {
		ret = types.Resolve(typ)
		if ret != nil {
			break
		}
	}
	return
}
