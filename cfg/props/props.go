package props

import (
	"regexp"
	"strings"
	"bytes"
	"errors"
	"os"
	"reflect"
)

var linePattern, _ = regexp.Compile("[#].*\\n|\\s+\\n|\\S+[=]|.*\n")

func Unmarshal(data bytes.Buffer, out interface{}) (err error) {
	v := reflect.ValueOf(out)
	if v.Kind() == reflect.Map && !v.IsNil() {
		_, err = ParseIntoMap(data, out.(map[string]string))
	} else {
		err = errors.New("Can't unmarshal properties on an custom object, only map are supported")
	}
	return
}

func Parse(data bytes.Buffer) (map[string]string, error) {
	return ParseIntoMap(data, make(map[string]string))
}

func ParseIntoMap(data bytes.Buffer, fill map[string]string) (ret map[string]string, err error) {
	ret = fill
	str := data.String()
	if !strings.HasSuffix(str, "\n") {
		str = str + "\n"
	}
	s2 := linePattern.FindAllString(str, -1)

	for i := 0; i < len(s2); {
		if strings.HasPrefix(s2[i], "#") {
			i++
		} else if strings.HasSuffix(s2[i], "=") {
			key := s2[i][0: len(s2[i])-1]
			i++
			if strings.HasSuffix(s2[i], "\n") {
				val := s2[i][0: len(s2[i])-1]
				if strings.HasSuffix(val, "\r") {
					val = val[0: len(val)-1]
				}
				i++
				fill[key] = val
			}
		} else if strings.Index(" \t\r\n", s2[i][0:1]) > -1 {
			i++
		} else {
			err = errors.New("Unable to process line in cfg file containing " + s2[i])
			break
		}
	}
	return
}

func Environ() map[string]string {
	return EnvironIntoMap(make(map[string]string))
}

func EnvironIntoMap(fill map[string]string) (ret map[string]string) {
	ret = fill
	SplitPropertiesIntoMap(os.Environ(), ret)
	return
}

func SplitPropertiesIntoMap(params []string, fill map[string]string) {
	for _, e := range params {
		pair := strings.Split(e, "=")
		fill[pair[0]] = pair[1]
	}
}
