package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/valid"
)

func validate(in valid.Validatable) error {
	if p := in.Validate(); !p.Ok() {
		return service.Validation(p)
	}
	return nil
}

// parseIndex parses a positional integer argument (index / id).
func parseIndex(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, service.BadRequestf("invalid integer value %q", s)
	}
	return i, nil
}

// parseHeaders turns "Name: value" flag strings into a header map.
func parseHeaders(headers []string) (map[string][]string, error) {
	hh := make(map[string][]string)
	for _, header := range headers {
		if !strings.Contains(header, ":") {
			return nil, service.BadRequestf("header %q must contain \":\"", header)
		}
		parts := strings.SplitN(header, ":", 2)
		name, value := parts[0], strings.TrimLeft(parts[1], " ")
		hh[name] = append(hh[name], value)
	}
	return hh, nil
}

// readBody returns the route body: either the raw argument or, when file is set,
// the contents of the file it names.
func readBody(arg string, file bool) ([]byte, error) {
	if file {
		b, err := os.ReadFile(arg)
		if err != nil {
			return nil, service.BadRequestf("fail to read file %q", arg)
		}
		return b, nil
	}
	return []byte(arg), nil
}
