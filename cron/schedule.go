package cron

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type bounds struct {
	min, max uint
}

// The bounds for each field.
var (
	minutes = bounds{0, 59}
	hours   = bounds{0, 23}
	dom     = bounds{1, 31}
	months  = bounds{1, 12}
	dow     = bounds{0, 6}
)

type Schedule struct {
	Minute, Hour, Day, Month, WeekDay uint64
}

const (
	starBit = 1 << 63
)

func (s *Schedule) Next(t time.Time) time.Time {
	// 找到下一个执行时间
	t = t.Add(time.Minute - time.Duration(t.Nanosecond()))
	added := false

	// If no time is found within five years, return zero.
	yearLimit := t.Year() + 5

WRAP:
	if t.Year() > yearLimit {
		return time.Time{}
	}

	// Find the first applicable month.
	// If it's this month, then do nothing.
	for 1<<uint(t.Month())&s.Month == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 1, 0)

		// 第二年，从头开始匹配合适的时间
		if t.Month() == time.January {
			goto WRAP
		}
	}

	// Now get a day in that month.
	for !dayMatches(s, t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 0, 1)

		if t.Day() == 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		}
		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	return t
}

func dayMatches(s *Schedule, t time.Time) bool {
	var (
		dayMatch     bool = 1<<uint(t.Day())&s.Day > 0
		weekDayMatch bool = 1<<uint(t.Weekday())&s.WeekDay > 0
	)
	if s.Day&starBit > 0 || s.WeekDay&starBit > 0 {
		return dayMatch && weekDayMatch
	}
	return dayMatch || weekDayMatch
}

// * 30 5 * * *
func Parse(cronExpr string) (*Schedule, error) {
	if len(cronExpr) == 0 {
		return nil, fmt.Errorf("Empty cronExpr")
	}

	fields := strings.Fields(cronExpr)

	if len(fields) != 5 {
		return nil, fmt.Errorf("Wrong cron expression")
	}

	var err error
	fieldFunc := func(field string, r bounds) uint64 {
		if err != nil {
			return 0
		}
		var bits uint64
		bits, err = getField(field, r)
		return bits
	}
	schedule := &Schedule{
		Minute:  fieldFunc(fields[0], minutes),
		Hour:    fieldFunc(fields[1], hours),
		Day:     fieldFunc(fields[2], dom),
		Month:   fieldFunc(fields[3], months),
		WeekDay: fieldFunc(fields[4], dow),
	}
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

// getField returns an Int with the bits set representing all of the times that
// the field represents or error parsing field value.  A "field" is a comma-separated
// list of "ranges".
func getField(field string, r bounds) (uint64, error) {
	var bits uint64
	ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	for _, expr := range ranges {
		bit, err := getRange(expr, r)
		if err != nil {
			return bits, err
		}
		bits |= bit
	}
	return bits, nil
}

// getRange returns the bits indicated by the given expression:
//   number | number "-" number [ "/" number ]
// or error parsing range.
func getRange(expr string, r bounds) (uint64, error) {
	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
		err              error
	)

	var extra uint64
	if lowAndHigh[0] == "*" {
		start = r.min
		end = r.max
		extra = starBit
	} else {
		start, err = ParseInt(lowAndHigh[0])
		if err != nil {
			return 0, err
		}
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			end, err = ParseInt(lowAndHigh[1])
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("Too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = ParseInt(rangeAndStep[1])
		if err != nil {
			return 0, err
		}
		if singleDigit {
			end = r.max
		}
	default:
		return 0, fmt.Errorf("Too many slashes: %s", expr)
	}

	if start < r.min {
		return 0, fmt.Errorf("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		return 0, fmt.Errorf("End of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		return 0, fmt.Errorf("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}
	if step == 0 {
		return 0, fmt.Errorf("Step of range should be a positive number: %s", expr)
	}

	return getBits(start, end, step) | extra, nil
}

func ParseInt(expr string) (uint, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		return 0, fmt.Errorf("Negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num), nil
}

func getBits(min, max, step uint) uint64 {
	var bits uint64

	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

// all returns all bits within the given bounds.  (plus the star bit)
func all(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}
