package xml

import (
	"encoding/xml"
)

func FindAttr(attrs []xml.Attr, space, local string) (ret *xml.Attr) {
	for _, attr := range attrs {
		if attr.Name.Local == local && attr.Name.Space == space {
			ret = &attr
			break
		}
	}
	return
}
