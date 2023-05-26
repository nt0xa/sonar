package database

import "time"

func now() time.Time {
	// PostgreSQL currently can't store nanoseconds
	// https://github.com/lib/pq/issues/227
	return time.Now().UTC().Truncate(time.Microsecond)
}
