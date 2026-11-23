// Package mini_business_day computes working-day arithmetic by skipping weekends
// and a caller-supplied list of holiday dates.
//
// The package deliberately treats holidays as a flat set of dates rather than
// trying to model recurring rules. That keeps the API small and makes the
// behaviour of the calculator fully determined by its inputs.
package mini_business_day

import (
	"errors"
	"sort"
	"time"
)

// Calendar holds a set of holiday dates. Dates are normalised to UTC midnight
// so that equality checks are stable regardless of the time component.
type Calendar struct {
	holidays map[time.Time]struct{}
}

// NewCalendar returns a Calendar containing the supplied holiday dates.
//
// Dates are truncated to UTC midnight before being stored. Duplicate dates are
// collapsed, which means passing the same holiday twice is harmless.
func NewCalendar(holidays ...time.Time) *Calendar {
	c := &Calendar{holidays: make(map[time.Time]struct{}, len(holidays))}
	for _, h := range holidays {
		c.holidays[normalize(h)] = struct{}{}
	}
	return c
}

// IsHoliday reports whether the given date is in the calendar. The time
// component is ignored.
func (c *Calendar) IsHoliday(t time.Time) bool {
	if c == nil {
		return false
	}
	_, ok := c.holidays[normalize(t)]
	return ok
}

// IsBusinessDay reports whether t falls on a weekday and is not a holiday.
func (c *Calendar) IsBusinessDay(t time.Time) bool {
	d := normalize(t)
	if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		return false
	}
	return !c.IsHoliday(d)
}

// AddBusinessDays returns the date n business days after start. If n is
// negative the calculation moves backwards. The returned time is always at
// UTC midnight.
//
// A start date that is itself a non-business day is first rolled forward (for
// positive n) or backward (for negative n) to the nearest business day. This
// matches the common expectation that "add 1 business day to Saturday" gives
// Monday rather than Tuesday.
func (c *Calendar) AddBusinessDays(start time.Time, n int) (time.Time, error) {
	if n == 0 {
		return normalize(start), nil
	}

	step := 1
	if n < 0 {
		step = -1
		n = -n
	}

	cur := normalize(start)
	// Roll onto a business day first, so that the count is measured from the
	// next actual working day.
	for !c.IsBusinessDay(cur) {
		cur = cur.AddDate(0, 0, step)
	}

	for n > 0 {
		cur = cur.AddDate(0, 0, step)
		if c.IsBusinessDay(cur) {
			n--
		}
	}
	return cur, nil
}

// BusinessDaysBetween returns the number of business days in the half-open
// interval [start, end). The result is negative when end is before start.
//
// The calculation is symmetric: swapping the arguments negates the result.
func (c *Calendar) BusinessDaysBetween(start, end time.Time) int {
	s := normalize(start)
	e := normalize(end)
	if s.Equal(e) {
		return 0
	}

	sign := 1
	if e.Before(s) {
		sign = -1
		s, e = e, s
	}

	count := 0
	for d := s; d.Before(e); d = d.AddDate(0, 0, 1) {
		if c.IsBusinessDay(d) {
			count++
		}
	}
	return sign * count
}

// NextBusinessDay returns the first business day strictly after t. It returns
// an error if no business day can be found within a reasonable search window,
// which can only happen if every day in that window is a holiday or weekend.
func (c *Calendar) NextBusinessDay(t time.Time) (time.Time, error) {
	return c.seek(t, 1)
}

// PrevBusinessDay returns the first business day strictly before t.
func (c *Calendar) PrevBusinessDay(t time.Time) (time.Time, error) {
	return c.seek(t, -1)
}

// seek walks in the given direction until it finds a business day. The bound
// exists purely to avoid an infinite loop if a caller builds a pathological
// calendar; 3660 days is a decade, far beyond any realistic holiday list.
func (c *Calendar) seek(t time.Time, step int) (time.Time, error) {
	cur := normalize(t)
	for i := 0; i < 3660; i++ {
		cur = cur.AddDate(0, 0, step)
		if c.IsBusinessDay(cur) {
			return cur, nil
		}
	}
	return time.Time{}, errors.New("mini_business_day: no business day found within search window")
}

// Holidays returns a sorted copy of the holiday dates in the calendar. The
// copy prevents callers from mutating the internal set.
func (c *Calendar) Holidays() []time.Time {
	if c == nil || len(c.holidays) == 0 {
		return nil
	}
	out := make([]time.Time, 0, len(c.holidays))
	for d := range c.holidays {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

// normalize strips the time-of-day and location so that two instants on the
// same calendar day compare equal.
func normalize(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
