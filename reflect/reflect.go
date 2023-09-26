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
	LabeledTypes map[string]*Types
}

func (o *LabeledTypes) Register(label string) (ret *Types) {
	ret = &Types{
		Label: label,
		Types: make(map[string]reflect.Type),
	}
	o.LabeledTypes[label] = ret
	return
}

func (o *LabeledTypes) Resolve(namespace, name string) (ret reflect.Type) {
	if namespaceTypes := o.LabeledTypes[namespace]; namespaceTypes != nil {
		ret = namespaceTypes.Resolve(name)
	}
	return
}
