package templates

import (
	htmltemplate "html/template"
)

var defaultOptions = options{
	html:    true,
	newLine: true,
	markup: map[string]string{
		"<bold>":  "",
		"</bold>": "",
		"<code>":  "",
		"</code>": "",
		"<pre>":   "",
		"</pre>":  "",
	},
}

type options struct {
	markup     map[string]string
	extraFuncs htmltemplate.FuncMap
	html       bool
	newLine    bool
}

type Option func(*options)

func NewLine(b bool) Option {
	return func(opts *options) {
		opts.newLine = b
	}
}

func HTMLEscape(b bool) Option {
	return func(opts *options) {
		opts.html = b
	}
}

func Markup(mopts ...MarkupOption) Option {
	return func(opts *options) {
		for _, mopt := range mopts {
			mopt(opts.markup)
		}
	}
}

type MarkupOption func(markup map[string]string)

func Bold(open, close string) MarkupOption {
	return func(markup map[string]string) {
		markup["<bold>"] = open
		markup["</bold>"] = close
	}
}

func CodeBlock(open, close string) MarkupOption {
	return func(markup map[string]string) {
		markup["<pre>"] = open
		markup["</pre>"] = close
	}
}

func CodeInline(open, close string) MarkupOption {
	return func(markup map[string]string) {
		markup["<code>"] = open
		markup["</code>"] = close
	}
}
