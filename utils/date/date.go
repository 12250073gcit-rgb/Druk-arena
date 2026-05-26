package date

import "time"

// Today returns today's date as a string (YYYY-MM-DD)
func Today() string {
	return time.Now().Format("2006-01-02")
}

// FormatDate formats a time.Time to YYYY-MM-DD string
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
