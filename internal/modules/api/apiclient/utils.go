package apiclient

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/fatih/structs"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/schema"
)

var encoder = schema.NewEncoder()

func init() {
	encoder.SetAliasTag("query")
}

func toQuery(src interface{}) url.Values {
	dst := url.Values{}
	if err := encoder.Encode(src, dst); err != nil {
		panic(err)
	}
	return dst
}

func toPath(src interface{}) map[string]string {
	dst := make(map[string]string)

	for _, f := range structs.Fields(src) {
		if f.Tag("path") == "" {
			continue
		}

		dst[f.Tag("path")] = fmt.Sprintf("%s", f.Value())
	}

	return dst
}

func handle(resp *resty.Response, err error) errors.Error {
	if err != nil {
		return errors.Internal(err)
	}

	if resp.IsError() &&
		!strings.Contains(resp.Header().Get("Content-Type"), "application/json") {
		return errors.Internalf(resp.String())
	}

	if resp.Error() != nil {
		return resp.Error().(*APIError)
	}

	return nil
}
