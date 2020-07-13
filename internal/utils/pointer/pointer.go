package pointer

import "time"

func Int64(v int64) *int64 {
	return &v
}

func String(v string) *string {
	return &v
}

func Time(v time.Time) *time.Time {
	return &v
}
