package valid

import "fmt"

// --- Max ---

type maxRule struct{ n int64 }

func (r maxRule) checkNumber(v int64) string {
	if v > r.n {
		return fmt.Sprintf("must be no greater than %d", r.n)
	}
	return ""
}

// Max asserts the number is no greater than n.
func Max(n int64) NumberRule { return maxRule{n} }
