package eio

import (
	"bytes"
	"io"
	"os"
	"strings"
)

type ReverseScanner struct {
	r              io.ReaderAt
	pos            int
	err            error
	buf            []byte
	skipEmptyLines bool

	// last scan
	scanBytes []byte
	scanStart int
	scanErr   error
}

func NewReverseScanner(r io.ReaderAt, pos int) *ReverseScanner {
	return &ReverseScanner{r: r, pos: pos, skipEmptyLines: true}
}

func NewReverseScannerString(text string) *ReverseScanner {
	return NewReverseScanner(strings.NewReader(text), len(text))
}

func NewReverseScannerFile(file *os.File) (*ReverseScanner, error) {
	if fi, err := file.Stat(); err != nil {
		return nil, err
	} else {
		return NewReverseScanner(file, int(fi.Size())), nil
	}
}

func (s *ReverseScanner) readMore() {
	if s.pos == 0 {
		s.err = io.EOF
		return
	}
	size := 1024
	if size > s.pos {
		size = s.pos
	}
	s.pos -= size
	buf2 := make([]byte, size, size+len(s.buf))

	// ReadAt attempts to read full buff!
	_, s.err = s.r.ReadAt(buf2, int64(s.pos))
	if s.err == nil {
		s.buf = append(buf2, s.buf...)
	}
}
func (s *ReverseScanner) Scan() bool {
	s.scanBytes, s.scanStart, s.scanErr = s.line()
	if s.scanErr == nil && len(s.scanBytes) == 0 {
		s.Scan()
	}
	return s.scanErr == nil
}
func (s *ReverseScanner) Bytes() []byte {
	return s.scanBytes
}

func (s *ReverseScanner) Text() string {
	return string(s.scanBytes)
}

func (s *ReverseScanner) ScanErr() error {
	return s.scanErr
}

func (s *ReverseScanner) BytesStartErr() (line []byte, start int, err error) {
	return s.scanBytes, s.scanStart, s.scanErr
}

func (s *ReverseScanner) line() (line []byte, start int, err error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	for {
		lineStart := bytes.LastIndexByte(s.buf, '\n')
		if lineStart >= 0 {
			// We have a complete line:
			var line []byte
			line, s.buf = dropCR(s.buf[lineStart+1:]), s.buf[:lineStart]
			return line, s.pos + lineStart + 1, nil
		}
		// Need more data:
		s.readMore()
		if s.err != nil {
			if s.err == io.EOF {
				if len(s.buf) > 0 {
					return dropCR(s.buf), 0, nil
				}
			}
			return nil, 0, s.err
		}
	}
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
