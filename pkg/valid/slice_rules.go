package valid

import "fmt"

// --- MinItems / MaxItems ---

type minItemsRule struct{ n int }

func (r minItemsRule) msg() string { return fmt.Sprintf("must contain at least %d items", r.n) }
func (r minItemsRule) checkSlice(n int) string {
	if n < r.n {
		return r.msg()
	}
	return ""
}
func (r minItemsRule) checkStringSlice(e []string) string { return r.checkSlice(len(e)) }

// MinItems asserts the slice has at least n elements.
func MinItems(n int) LengthRule { return minItemsRule{n} }

type maxItemsRule struct{ n int }

func (r maxItemsRule) msg() string { return fmt.Sprintf("must contain no more than %d items", r.n) }
func (r maxItemsRule) checkSlice(n int) string {
	if n > r.n {
		return r.msg()
	}
	return ""
}
func (r maxItemsRule) checkStringSlice(e []string) string { return r.checkSlice(len(e)) }

// MaxItems asserts the slice has no more than n elements.
func MaxItems(n int) LengthRule { return maxItemsRule{n} }

// --- Each ---

type eachRule struct{ rules []StringRule }

func (r eachRule) checkStringSlice(e []string) string {
	for i, s := range e {
		if msg := applyString(s, r.rules); msg != "" {
			return fmt.Sprintf("element #%d: %s", i, msg)
		}
	}
	return ""
}

// Each applies string rules to every element of a string slice.
func Each(rules ...StringRule) StringSliceRule { return eachRule{rules} }
