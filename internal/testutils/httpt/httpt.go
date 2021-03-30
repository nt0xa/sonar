package httpt

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// Body fields
//

type stringFormField struct {
	s string
}

func (f *stringFormField) String() string {
	return f.s
}

func (f *stringFormField) Reader() io.Reader {
	return strings.NewReader(f.s)
}

func (f *stringFormField) Writer(w *multipart.Writer, name string) (io.Writer, error) {
	fw, err := w.CreateFormField(name)
	if err != nil {
		return nil, err
	}

	return fw, nil
}

type fileFormField struct {
	Name  string
	inner *stringFormField
}

func (f *fileFormField) String() string {
	return f.inner.String()
}

func (f *fileFormField) Reader() io.Reader {
	return f.inner.Reader()
}

func (f *fileFormField) Writer(w *multipart.Writer, name string) (io.Writer, error) {
	fw, err := w.CreateFormFile(name, f.Name)
	if err != nil {
		return nil, err
	}

	return fw, nil
}

type FormField interface {
	String() string
	Reader() io.Reader
	Writer(*multipart.Writer, string) (io.Writer, error)
}

func StringField(s string) FormField {
	return &stringFormField{s}
}

func FileField(name string, data string) FormField {
	return &fileFormField{name, &stringFormField{data}}
}

//
// Response matchers
//

type Matcher func(*testing.T, interface{})

func Regex(re *regexp.Regexp) Matcher {
	return func(t *testing.T, value interface{}) {
		assert.Regexp(t, re, value)
	}
}

func Equal(expected interface{}) Matcher {
	return func(t *testing.T, value interface{}) {
		assert.EqualValues(t, expected, value)
	}
}

func Contains(s interface{}) Matcher {
	return func(t *testing.T, value interface{}) {
		require.NotNil(t, value)
		assert.Contains(t, value, s)
	}
}

type ResponseMatcher func(*testing.T, *http.Response)

func Code(c int) ResponseMatcher {
	return func(t *testing.T, r *http.Response) {
		assert.Equal(t, c, r.StatusCode)
	}
}

func Header(key string, match Matcher) ResponseMatcher {
	return func(t *testing.T, r *http.Response) {
		header := r.Header.Get(key)
		assert.NotEmpty(t, header)
		match(t, header)
	}
}

func Body(match Matcher) ResponseMatcher {
	return func(t *testing.T, r *http.Response) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		match(t, string(body))
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
}
