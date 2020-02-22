package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"unicode"
)

func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func HexDump(by []byte) string {
	s := ""
	n := len(by)
	rowcount := 0
	stop := (n / 8) * 8
	k := 0
	for i := 0; i <= stop; i += 8 {
		k++
		if i+8 < n {
			rowcount = 8
		} else {
			rowcount = Min(k*8, n) % 8
		}

		s += fmt.Sprintf("%02d:  ", i)
		for j := 0; j < rowcount; j++ {
			s += fmt.Sprintf("%02x  ", by[i+j])
		}
		for j := rowcount; j < 8; j++ {
			s += "    "
		}
		s += fmt.Sprintf("  %s\n", ViewString(by[i:(i+rowcount)]))
	}

	return s
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ViewString(b []byte) string {
	r := []rune(string(b))
	for i := range r {
		if r[i] > unicode.MaxASCII || !unicode.IsPrint(r[i]) {
			r[i] = '.'
		}
	}
	return string(r)
}

func StringPrintable(s string) bool {
	r := []rune(s)
	for i := range r {
		if r[i] > unicode.MaxASCII ||
			!(unicode.IsPrint(r[i]) || unicode.IsSpace(r[i])) {
			return false
		}
	}
	return true
}
