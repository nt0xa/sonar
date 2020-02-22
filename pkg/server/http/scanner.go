package http

import (
	"bytes"
	"regexp"
	"strconv"
)

var (
	contentLengthRegexp = regexp.MustCompile(`(?im)^Content-Length:\s+(\d+)\s+$`)
)

type HTTPScanner struct {
	body   bool
	length int
	end    bool
}

func (s *HTTPScanner) Scan(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if s.end {
		return 0, nil, ErrNoMoreData
	}

	if !s.body && contentLengthRegexp.Match(data) {
		// TODO: last header
		m := contentLengthRegexp.FindAllStringSubmatch(string(data), 1)
		l, _ := strconv.ParseUint(m[0][1], 10, 64)
		s.body = true
		s.length = int(l)
	}

	i := bytes.Index(data, []byte("\r\n\r\n"))
	if (s.length == 0 && i >= 0) ||
		(s.length > 0 && len(data[i+4:]) >= s.length) {
		s.end = true
		return len(data), data, nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}
