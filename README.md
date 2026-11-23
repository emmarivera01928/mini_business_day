# mini_business_day

A small Go library for working-day arithmetic that skips weekends and a caller-provided list of holiday dates.

```go
package main

import (
	"fmt"
	"time"

	"mini_business_day"
)

func main() {
	cal := mini_business_day.NewCalendar(
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
	)

	start := time.Date(2024, time.January, 5, 0, 0, 0, 0, time.UTC) // Friday
	end, err := cal.AddBusinessDays(start, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(end.Format("2006-01-02")) // 2024-01-08 (Monday)

	fmt.Println(cal.BusinessDaysBetween(start, end)) // 1
}
```

## Why this exists

Most date libraries either know nothing about business days or require a heavyweight calendar package with timezone and recurrence rules. This library takes the opposite position: holidays are just a flat set of dates supplied by the caller. The trade-off is that recurring holidays (e.g. "the first Monday in September") must be expanded by the caller each year, but in exchange the calculator has no hidden state, no configuration files, and no dependency beyond the standard library. That makes its behaviour completely determined by its inputs, which is what you want when the result feeds into a payroll or scheduling system.

## Edge cases

If you pass a start date that is itself a non-business day, `AddBusinessDays` first rolls it to the nearest business day in the direction of travel before counting. So adding one business day to a Saturday gives Monday, not Tuesday. `BusinessDaysBetween` uses a half-open interval `[start, end)` and returns a negative number when `end` is before `start`.

## API

- `NewCalendar(holidays ...time.Time) *Calendar`
- `(*Calendar) IsHoliday(t time.Time) bool`
- `(*Calendar) IsBusinessDay(t time.Time) bool`
- `(*Calendar) AddBusinessDays(start time.Time, n int) (time.Time, error)`
- `(*Calendar) BusinessDaysBetween(start, end time.Time) int`
- `(*Calendar) NextBusinessDay(t time.Time) (time.Time, error)`
- `(*Calendar) PrevBusinessDay(t time.Time) (time.Time, error)`
- `(*Calendar) Holidays() []time.Time`

All dates are normalised to UTC midnight internally, so the time-of-day component of an argument is ignored.
