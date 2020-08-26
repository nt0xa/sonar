package apiclient

import (
	"fmt"
	"net/url"

	"github.com/fatih/structs"
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
