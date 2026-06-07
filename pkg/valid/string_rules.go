package valid

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// --- Optional (marker) ---

type optionalRule struct{}

func (optionalRule) checkString(string) string { return "" }

// Optional skips the remaining string rules when the value is empty. Place it
// first to make a constrained field accept an empty value.
var Optional = optionalRule{}

// --- MinLength ---

type minLengthRule struct{ n int }

func (r minLengthRule) checkString(s string) string {
	if len([]rune(s)) < r.n {
		return fmt.Sprintf("must be at least %d characters", r.n)
	}
	return ""
}

// MinLength asserts the string has at least n characters.
func MinLength(n int) StringRule { return minLengthRule{n} }

// --- In ---

type inRule struct{ allowed []string }

func (r inRule) checkString(s string) string {
	if slices.Contains(r.allowed, s) {
		return ""
	}
	return fmt.Sprintf("must be one of: %s", strings.Join(r.allowed, ", "))
}

// In asserts the string equals one of the allowed values. It accepts ~string
// values (e.g. enum value slices via In(TypeValues()...)).
func In[T ~string](vals ...T) StringRule {
	allowed := make([]string, len(vals))
	for i, v := range vals {
		allowed[i] = string(v)
	}
	return inRule{allowed}
}

// --- Match ---

type matchRule struct {
	re  *regexp.Regexp
	msg string
}

func (r matchRule) checkString(s string) string {
	if !r.re.MatchString(s) {
		return r.msg
	}
	return ""
}

// Match asserts the string matches re, reporting msg on failure.
func Match(re *regexp.Regexp, msg string) StringRule { return matchRule{re, msg} }

// --- By ---

type byRule struct{ fn func(string) error }

func (r byRule) checkString(s string) string {
	if err := r.fn(s); err != nil {
		return err.Error()
	}
	return ""
}

// By adapts a func(string) error validator (e.g. a domain-specific check) into a
// rule.
func By(fn func(string) error) StringRule { return byRule{fn} }
