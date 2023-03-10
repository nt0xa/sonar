package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/russtone/sonar/internal/actions"
)

type Options struct {
	Type string
}

var options = &Options{}

func init() {
	flag.StringVar(&options.Type, "type", "", "Type (cmd, api, apiclient)")
}

type Action struct {
	Name     string
	Resource string
	Verb     string
	Params   struct {
		Name  string
		Types []string
	}
	Result     string
	HTTPMethod string
	HTTPPath   string
}

var tags = map[string]string{
	"json":  "JSON",
	"path":  "Path",
	"query": "Query",
}

func main() {
	flag.Parse()

	if !contains([]string{"cmd", "api", "apiclient"}, options.Type) {
		fmt.Fprintf(os.Stderr, "invalid type\n")
		os.Exit(1)
	}

	// Collect actions info.
	acts := make([]*Action, 0)
	t := reflect.TypeOf((*actions.Actions)(nil)).Elem()

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		act := &Action{Name: m.Name}

		// HTTP Method (for API and API client).
		if strings.Contains(m.Name, "Create") {
			act.HTTPMethod = "POST"
			act.Resource = strings.Replace(m.Name, "Create", "", 1)
			act.Verb = "Create"
		} else if strings.Contains(m.Name, "Update") {
			act.HTTPMethod = "PUT"
			act.Resource = strings.Replace(m.Name, "Update", "", 1)
			act.Verb = "Update"
		} else if strings.Contains(m.Name, "Delete") {
			act.HTTPMethod = "DELETE"
			act.Resource = strings.Replace(m.Name, "Delete", "", 1)
			act.Verb = "Delete"
		} else if strings.Contains(m.Name, "Get") {
			act.HTTPMethod = "GET"
			act.Resource = strings.Replace(m.Name, "Get", "", 1)
			act.Verb = "Get"
		} else if strings.Contains(m.Name, "List") {
			act.HTTPMethod = "GET"
			act.Resource = strings.Replace(m.Name, "List", "", 1)
			act.Verb = "List"
		} else {
			fmt.Fprintf(os.Stderr, "invalid name: %q\n", m.Name)
			os.Exit(1)
		}

		// HTTP path (with paramters).
		act.HTTPPath += "/" + strings.ToLower(pathName(
			regexp.MustCompile(`^[A-Z]+[a-z]+`).FindString(m.Name),
		))

		// Actions arguments.
		for j := 0; j < m.Type.NumIn(); j++ {

			// We only need *Parameters arg.
			arg := m.Type.In(j)
			if !strings.Contains(arg.Name(), "Params") {
				continue
			}

			act.Params.Name = arg.Name()

			// Iterate parameters fields and save which
			// of them came from path, query and json.
			for k := 0; k < arg.NumField(); k++ {
				f := arg.Field(k)

				for tag, typ := range tags {
					if f.Tag.Get(tag) != "" && !contains(act.Params.Types, typ) {
						act.Params.Types = append(act.Params.Types, typ)
					}
				}

				// Save path parameters in HTTPPath.
				if path := f.Tag.Get("path"); path != "" {
					act.HTTPPath += fmt.Sprintf("/{%s}", path)
				}
			}
		}

		// Action result.
		for j := 0; j < m.Type.NumOut(); j++ {
			res := m.Type.Out(j)

			// We only need *Result return type.
			if !strings.Contains(res.Name(), "Result") {
				continue
			}

			act.Result = res.Name()
		}

		acts = append(acts, act)
	}

	var code string

	switch options.Type {
	case "cmd":
		code = cmdCode
	case "api":
		code = apiCode
	case "apiclient":
		code = apiClientCode
	default:
		panic("must not happen")
	}

	// Render templates.
	tpl := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(code))

	if err := tpl.Execute(os.Stdout, acts); err != nil {
		fmt.Fprintf(os.Stderr, "template execution failed: %v", err)
	}
}

func contains(items []string, item string) bool {
	for _, it := range items {
		if it == item {
			return true
		}
	}

	return false
}

// DNSRecords -> DNS-Records
// Users -> Users
func pathName(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] > 'A' && s[i] < 'Z' {
			continue
		}

		if i <= 1 {
			break
		}

		return s[:i-1] + "-" + s[i-1:]
	}

	return s
}
