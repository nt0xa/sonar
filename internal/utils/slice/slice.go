package slice

import "sort"

// StringsDedup returns slice of strings without duplicates.
func StringsDedup(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	sort.Strings(items)

	j := 0
	for i := 1; i < len(items); i++ {
		if items[j] == items[i] {
			continue
		}

		j++

		items[j] = items[i]
	}

	return items[:j+1]
}

func StringsContains(items []string, item string) bool {
	for _, s := range items {
		if s == item {
			return true
		}
	}
	return false
}

func FindIndex(values []string, value string) int {
	for i, v := range values {
		if value == v {
			return i
		}
	}

	return -1
}
