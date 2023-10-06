package xml

import (
	"encoding/xml"
	"github.com/go-ee/utils/reflect"
)

func NewNamespacesTypes() *NamespacesTypes {
	return &NamespacesTypes{
		LabeledTypes: reflect.NewLabeledTypes[interface{}, *xml.Name](),
	}
}

type NamespacesTypes struct {
	*reflect.LabeledTypes[interface{}, *xml.Name]
}

func FindAttrForLocal(attrs []xml.Attr, local string) (ret *xml.Attr) {
	for _, attr := range attrs {
		if attr.Name.Local == local {
			ret = &attr
			break
		}
	}
	return
}

func FindAttrForSpaceAndLocal(attrs []xml.Attr, space, local string) (ret *xml.Attr) {
	for _, attr := range attrs {
		if attr.Name.Local == local && attr.Name.Space == space {
			ret = &attr
			break
		}
	}
	return
}
