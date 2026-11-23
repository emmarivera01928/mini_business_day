package mini_business_day

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, year int, month time.Month, day int) time.Time {
	t.Helper()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestIsBusinessDay(t *testing.T) {
	cal := NewCalendar(mustDate(t, 2024, time.January, 1))

	cases := []struct {
		name string
		date time.Time
		want bool
	}{
		{"ordinary weekday", mustDate(t, 2024, time.January, 2), true},
		{"saturday", mustDate(t, 2024, time.January, 6), false},
		{"sunday", mustDate(t, 2024, time.January, 7), false},
		{"holiday", mustDate(t, 2024, time.January, 1), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cal.IsBusinessDay(tc.date); got != tc.want {
				t.Fatalf("IsBusinessDay(%v) = %v, want %v", tc.date, got, tc.want)
			}
		})
	}
}

func TestIsBusinessDayIgnoresTimeOfDay(t *testing.T) {
	cal := NewCalendar()
	morning := time.Date(2024, time.January, 8, 9, 30, 0, 0, time.UTC)
	evening := time.Date(2024, time.January, 8, 23, 59, 0, 0, time.UTC)
	if !cal.IsBusinessDay(morning) || !cal.IsBusinessDay(evening) {
		t.Fatal("expected both morning and evening on a Monday to be business days")
	}
}

func TestAddBusinessDaysForward(t *testing.T) {
	cal := NewCalendar(mustDate(t, 2024, time.January, 1))

	cases := []struct {
		name  string
		start time.Time
		n     int
		want  time.Time
	}{
		{"zero returns start", mustDate(t, 2024, time.January, 3), 0, mustDate(t, 2024, time.January, 3)},
		{"one from monday", mustDate(t, 2024, time.January, 2), 1, mustDate(t, 2024, time.January, 3)},
		{"three crosses weekend", mustDate(t, 2024, time.January, 4), 3, mustDate(t, 2024, time.January, 9)},
		{"skip holiday", mustDate(t, 2023, time.December, 29), 1, mustDate(t, 2024, time.January, 2)},
		{"start on saturday rolls forward", mustDate(t, 2024, time.January, 6), 1, mustDate(t, 2024, time.January, 9)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cal.AddBusinessDays(tc.start, tc.n)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("AddBusinessDays(%v, %d) = %v, want %v", tc.start, tc.n, got, tc.want)
			}
		})
	}
}

func TestAddBusinessDaysBackward(t *testing.T) {
	cal := NewCalendar(mustDate(t, 2024, time.January, 1))

	got, err := cal.AddBusinessDays(mustDate(t, 2024, time.January, 3), -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := mustDate(t, 2024, time.January, 2)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	got, err = cal.AddBusinessDays(mustDate(t, 2024, time.January, 2), -2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = mustDate(t, 2023, time.December, 28)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBusinessDaysBetween(t *testing.T) {
	cal := NewCalendar(mustDate(t, 2024, time.January, 1))

	cases := []struct {
		name  string
		start time.Time
		end   time.Time
		want  int
	}{
		{"same day", mustDate(t, 2024, time.January, 2), mustDate(t, 2024, time.January, 2), 0},
		{"one business day", mustDate(t, 2024, time.January, 2), mustDate(t, 2024, time.January, 3), 1},
		{"across weekend", mustDate(t, 2024, time.January, 5), mustDate(t, 2024, time.January, 8), 1},
		{"with holiday", mustDate(t, 2023, time.December, 29), mustDate(t, 2024, time.January, 3), 2},
		{"reversed", mustDate(t, 2024, time.January, 3), mustDate(t, 2024, time.January, 2), -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cal.BusinessDaysBetween(tc.start, tc.end); got != tc.want {
				t.Fatalf("BusinessDaysBetween(%v, %v) = %d, want %d", tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestNextAndPrevBusinessDay(t *testing.T) {
	cal := NewCalendar(mustDate(t, 2024, time.January, 1))

	next, err := cal.NextBusinessDay(mustDate(t, 2023, time.December, 29))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := mustDate(t, 2024, time.January, 2); !next.Equal(want) {
		t.Fatalf("NextBusinessDay = %v, want %v", next, want)
	}

	prev, err := cal.PrevBusinessDay(mustDate(t, 2024, time.January, 2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := mustDate(t, 2023, time.December, 29); !prev.Equal(want) {
		t.Fatalf("PrevBusinessDay = %v, want %v", prev, want)
	}
}

func TestHolidaysReturnsSortedCopy(t *testing.T) {
	cal := NewCalendar(
		mustDate(t, 2024, time.January, 1),
		mustDate(t, 2023, time.December, 25),
	)

	got := cal.Holidays()
	if len(got) != 2 {
		t.Fatalf("len(Holidays) = %d, want 2", len(got))
	}
	if !got[0].Equal(mustDate(t, 2023, time.December, 25)) || !got[1].Equal(mustDate(t, 2024, time.January, 1)) {
		t.Fatalf("Holidays not sorted: %v", got)
	}

	// Mutating the returned slice must not affect the calendar.
	got[0] = mustDate(t, 2000, time.January, 1)
	if cal.IsHoliday(mustDate(t, 2000, time.January, 1)) {
		t.Fatal("mutating the returned slice changed the calendar")
	}
}

func TestNilCalendarIsUsable(t *testing.T) {
	var cal *Calendar
	if cal.IsHoliday(mustDate(t, 2024, time.January, 1)) {
		t.Fatal("nil calendar should not report holidays")
	}
	if !cal.IsBusinessDay(mustDate(t, 2024, time.January, 2)) {
		t.Fatal("nil calendar should treat weekdays as business days")
	}
}
