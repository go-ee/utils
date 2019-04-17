package eio

import (
	"io"
	"os"
	"regexp"
	"strconv"
)

type MapWriter interface {
	WriteMap(data map[string]interface{}) error
}

type WriteCloserMapWriter struct {
	Convert func(map[string]interface{}) (io.Reader, error)
	Out     io.WriteCloser
}

func (o *WriteCloserMapWriter) WriteMap(data map[string]interface{}) (err error) {
	var reader io.Reader
	if reader, err = o.Convert(data); err == nil {
		_, err = io.Copy(o.Out, reader)
	}
	return
}

type CollectMapWriter struct {
	Data []map[string]interface{}
}

func NewCollectMapWriter() *CollectMapWriter {
	return &CollectMapWriter{Data: make([]map[string]interface{}, 0)}
}

func (o *CollectMapWriter) WriteMap(data map[string]interface{}) (err error) {
	o.Data = append(o.Data, data)
	return
}

func JoinInt64(ns []int64, sep string) string {
	if len(ns) == 0 {
		return ""
	}

	// Appr. 3 chars per num plus the comma.
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	// Or simply
	//   b := []byte{}
	for _, n := range ns {
		b = strconv.AppendInt(b, n, 10)
		b = append(b, ',')
	}
	b = b[:len(b)-1]
	return string(b)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func compileFileNameReg() *regexp.Regexp {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg
}

var fileNameReg *regexp.Regexp = compileFileNameReg()

func ToValidFileName(name string) string {
	return fileNameReg.ReplaceAllString(name, "_")
}
