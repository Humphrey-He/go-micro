package timex

import (
	"time"
)

const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "2006-01-02 15:04:05"
	DateTimeFormat = time.RFC3339
)

// Format formats a time as date string.
func Format(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTime formats a time as datetime string.
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDateTime formats a time as RFC3339 datetime string.
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// ParseDate parses a date string.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// ParseDateTime parses a datetime string.
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(TimeFormat, s)
}

// ParseDateTimeWithLocation parses a datetime string with location.
func ParseDateTimeWithLocation(s, loc string) (time.Time, error) {
	location, err := time.LoadLocation(loc)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(TimeFormat, s, location)
}

// Now returns the current time.
func Now() time.Time {
	return time.Now()
}

// Today returns the start of today.
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// StartOfDay returns the start of the day.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// AddDays adds days to a time.
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths adds months to a time.
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// SubDays returns the number of days between two times.
func SubDays(t1, t2 time.Time) int {
	return int(t1.Sub(t2).Hours() / 24)
}

// WithinDays checks if t is within the specified days from now.
func WithinDays(t time.Time, days int) bool {
	threshold := time.Now().AddDate(0, 0, -days)
	return t.After(threshold)
}

// Age calculates the age from the birth date.
func Age(birthdate time.Time) int {
	now := time.Now()
	age := now.Year() - birthdate.Year()
	if now.YearDay() < birthdate.YearDay() {
		age--
	}
	return age
}

// StartOfWeek returns the start of the week (Monday).
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1)
}

// EndOfWeek returns the end of the week (Sunday).
func EndOfWeek(t time.Time) time.Time {
	start := StartOfWeek(t)
	return start.AddDate(0, 0, 6).Add(-time.Second)
}

// Quarter returns the quarter of the year (1-4).
func Quarter(t time.Time) int {
	return (int(t.Month())-1)/3 + 1
}

// StartOfQuarter returns the start of the quarter.
func StartOfQuarter(t time.Time) time.Time {
	quarter := Quarter(t)
	month := (quarter-1)*3 + 1
	return time.Date(t.Year(), time.Month(month), 1, 0, 0, 0, 0, t.Location())
}

// EndOfQuarter returns the end of the quarter.
func EndOfQuarter(t time.Time) time.Time {
	start := StartOfQuarter(t).AddDate(0, 3, 0)
	return start.Add(-time.Second)
}

// UnixMilli returns the Unix timestamp in milliseconds.
func UnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// UnixMicro returns the Unix timestamp in microseconds.
func UnixMicro(t time.Time) int64 {
	return t.UnixMicro()
}

// FromUnixMilli creates a time from Unix milliseconds.
func FromUnixMilli(ms int64) time.Time {
	return time.UnixMilli(ms)
}
