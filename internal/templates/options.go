package templates

import (
	texttemplate "text/template"
)

// options represent all template options: default + perTemplate.
type options struct {
	defaultOptions templateOptions
	perTemplate    map[string]templateOptions
}

// get returns template options for the provided template id.
// If there are no specific options for the id, default options will be returned.
func (opt *options) get(id string) templateOptions {
	opts, ok := opt.perTemplate[id]
	if ok {
		return opts
	}
	return opt.defaultOptions
}

// templateOptions represent template rendering options.
type templateOptions struct {
	markup     map[string]string
	html       bool
	newLine    bool
	extraFuncs texttemplate.FuncMap
}

// defaultTemplateOptions is the default value for template options.
func defaultTemplateOptions() templateOptions {
	return templateOptions{
		html:       true,
		newLine:    true,
		markup:     newMarkup(),
		extraFuncs: make(texttemplate.FuncMap),
	}
}

func newMarkup() map[string]string {
	return map[string]string{
		"<bold>":  "",
		"</bold>": "",
		"<code>":  "",
		"</code>": "",
		"<pre>":   "",
		"</pre>":  "",
	}
}

type Option func(*options)

// Default allows to modify default template options.
func Default(topts ...TemplateOption) Option {
	return func(opts *options) {
		for _, topt := range topts {
			topt(&opts.defaultOptions)
		}
	}
}

// PerTemplate allows to modify options for single templates by their id.
func PerTemplate(id string, topts ...TemplateOption) Option {
	return func(opts *options) {
		op, ok := opts.perTemplate[id]
		if !ok {
			op = defaultTemplateOptions()
			opts.perTemplate[id] = op
		}

		for _, topt := range topts {
			topt(&op)
		}
	}
}

type TemplateOption func(*templateOptions)

func NewLine(b bool) TemplateOption {
	return func(opts *templateOptions) {
		opts.newLine = b
	}
}

func HTMLEscape(b bool) TemplateOption {
	return func(opts *templateOptions) {
		opts.html = b
	}
}

func Markup(mopts ...MarkupOption) TemplateOption {
	return func(opts *templateOptions) {
		for _, mopt := range mopts {
			mopt(opts.markup)
		}
	}
}

func ExtraFunc(name string, fn any) TemplateOption {
	return func(opts *templateOptions) {
		opts.extraFuncs[name] = fn
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
