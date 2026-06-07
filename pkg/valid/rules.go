package valid

// Rule categories. A rule reports a problem message, or "" when the value is ok.
// The category interfaces are what make the API type-safe: a String field only
// accepts StringRule, an Int field only accepts NumberRule, a Slice only accepts
// SliceRule, and Each is a StringSliceRule so it cannot be applied to a
// non-string slice.
type (
	StringRule      interface{ checkString(string) string }
	NumberRule      interface{ checkNumber(int64) string }
	SliceRule       interface{ checkSlice(int) string }
	StringSliceRule interface{ checkStringSlice([]string) string }
)

// LengthRule is a slice length rule usable on both Slice and StringSlice.
type LengthRule interface {
	SliceRule
	StringSliceRule
}

// --- Required (cross-type) ---

type requiredRule struct{}

func (requiredRule) checkString(s string) string {
	if s == "" {
		return "cannot be blank"
	}
	return ""
}

func (requiredRule) checkNumber(n int64) string {
	if n == 0 {
		return "cannot be blank"
	}
	return ""
}

func (requiredRule) checkSlice(n int) string {
	if n == 0 {
		return "cannot be blank"
	}
	return ""
}

func (requiredRule) checkStringSlice(e []string) string {
	if len(e) == 0 {
		return "cannot be blank"
	}
	return ""
}

// Required asserts the value is not its zero value: a non-empty string, a
// non-zero number, or a non-empty slice. It works on any field category.
var Required = requiredRule{}
