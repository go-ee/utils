package cfg

import (
	"reflect"
	"path/filepath"
	"bytes"
	"strings"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/BurntSushi/toml"
	"errors"
	"time"
	"text/template"
	"github.com/eugeis/gee/cfg/props"
)

func UnmarshalFile(config interface{}, file string) (err error) {
	return Unmarshal(config, []string{file}, []string{})
}

func Unmarshal(config interface{}, files []string, fileSuffixes []string) (err error) {
	var envAndProps map[string]string
	envAndProps, err = LoadEnvAndProperties(files, fileSuffixes)

	for _, file := range files {
		if !isPropertiesFile(file) {
			if err = LoadConfig(config, file, fileSuffixes, envAndProps); err != nil {
				break
			}
		}
	}
	return
}

func isPropertiesFile(file string) bool {
	return strings.HasSuffix(file, ".properties")
}

func LoadEnvAndProperties(files []string, fileSuffixes []string) (ret map[string]string, err error) {
	ret = props.Environ()
	for _, file := range files {
		if isPropertiesFile(file) {
			for _, fileWithSuffix := range CollectFilesForSuffixes(file, fileSuffixes) {
				if err = loadPropertiesFileIntoMap(fileWithSuffix, ret); err != nil {
					break
				}
			}
		}
	}
	return
}

func CollectFilesForSuffixes(file string, fileSuffixes []string) (ret []string) {
	return []string{file}
}

func loadPropertiesFileIntoMap(file string, toFoll map[string]string) (err error) {
	var data bytes.Buffer
	if data, err = ReadFileBindToProperties(file, toFoll); err == nil {
		_, err = props.ParseIntoMap(data, toFoll)
	}
	return
}

func LoadConfig(config interface{}, file string, fileSuffixes []string, properties map[string]string) (err error) {
	for _, fileWithSuffix := range CollectFilesForSuffixes(file, fileSuffixes) {
		if err = loadConfig(config, fileWithSuffix, properties); err != nil {
			break
		}
	}
	return
}

func loadConfig(config interface{}, file string, properties map[string]string) (err error) {
	var data bytes.Buffer
	if data, err = ReadFileBindToProperties(file, properties); err == nil {
		switch {
		case strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml"):
			err = yaml.Unmarshal(data.Bytes(), config)
		case strings.HasSuffix(file, ".toml"):
			err = toml.Unmarshal(data.Bytes(), config)
		case strings.HasSuffix(file, ".json"):
			err = json.Unmarshal(data.Bytes(), config)
		case strings.HasSuffix(file, ".properties"):
			err = props.Unmarshal(data, config)
		default:
			if toml.Unmarshal(data.Bytes(), config) != nil {
				if json.Unmarshal(data.Bytes(), config) != nil {
					if yaml.Unmarshal(data.Bytes(), config) != nil {
						if props.Unmarshal(data, config) != nil {
							err = errors.New("failed to decode config")
						}
					}
				}
			}
		}
	}
	return
}

func ReadFileBindToProperties(file string, params map[string]string) (data bytes.Buffer, err error) {
	tmpl := template.New(filepath.Base(file)).Funcs(template.FuncMap{
		"default": dfault,
	})
	if tmpl, err = tmpl.ParseFiles(file); err == nil {
		err = tmpl.Execute(&data, params)
	}
	return
}

//from Hugo
func dfault(dflt interface{}, given ...interface{}) (interface{}, error) {
	// given is variadic because the following construct will not pass a piped
	// argument when the key is missing:  {{ index . "key" | default "foo" }}
	// The Go template will complain that we got 1 argument when we expectd 2.

	if len(given) == 0 {
		return dflt, nil
	}
	if len(given) != 1 {
		return nil, fmt.Errorf("wrong number of args for default: want 2 got %d", len(given)+1)
	}

	g := reflect.ValueOf(given[0])
	if !g.IsValid() {
		return dflt, nil
	}

	set := false

	switch g.Kind() {
	case reflect.Bool:
		set = true
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		set = g.Len() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		set = g.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		set = g.Uint() != 0
	case reflect.Float32, reflect.Float64:
		set = g.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		set = g.Complex() != 0
	case reflect.Struct:
		switch actual := given[0].(type) {
		case time.Time:
			set = !actual.IsZero()
		default:
			set = true
		}
	default:
		set = !g.IsNil()
	}

	if set {
		return given[0], nil
	}

	return dflt, nil
}
