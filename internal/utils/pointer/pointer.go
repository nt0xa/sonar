package pointer

import "time"

func Bool(v bool) *bool {
	return &v
}

func Int64(v int64) *int64 {
	return &v
}

func String(v string) *string {
	return &v
}

func Time(v time.Time) *time.Time {
	return &v
}
