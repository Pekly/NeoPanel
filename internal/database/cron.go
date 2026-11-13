package database

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cron is a validated five-field, minute-resolution UTC cron expression.
// Day-of-month and day-of-week follow standard cron OR semantics when both restricted.
type Cron struct {
	fields [5]map[int]bool
	wild   [5]bool
}

func ParseCron(expr string) (*Cron, error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return nil, fmt.Errorf("cron must have exactly five fields (minute hour day month weekday)")
	}
	c := &Cron{}
	mins := [5]int{0, 0, 1, 1, 0}
	maxs := [5]int{59, 23, 31, 12, 7}
	for i, field := range parts {
		c.fields[i] = make(map[int]bool)
		c.wild[i] = strings.HasPrefix(field, "*")
		for _, item := range strings.Split(field, ",") {
			stepParts := strings.Split(item, "/")
			if len(stepParts) > 2 {
				return nil, fmt.Errorf("invalid cron field %q", field)
			}
			step := 1
			if len(stepParts) == 2 {
				v, err := strconv.Atoi(stepParts[1])
				if err != nil || v < 1 || v > maxs[i]-mins[i]+1 {
					return nil, fmt.Errorf("invalid cron step %q", item)
				}
				step = v
			}
			lo, hi := mins[i], maxs[i]
			if stepParts[0] != "*" {
				rangeParts := strings.Split(stepParts[0], "-")
				if len(rangeParts) > 2 {
					return nil, fmt.Errorf("invalid cron range %q", item)
				}
				var err error
				lo, err = cronNumber(rangeParts[0], i)
				if err != nil {
					return nil, err
				}
				hi = lo
				if len(rangeParts) == 2 {
					hi, err = cronNumber(rangeParts[1], i)
				} else if len(stepParts) == 2 {
					hi = maxs[i]
				}
				if err != nil || lo < mins[i] || hi > maxs[i] || hi < lo {
					return nil, fmt.Errorf("cron field out of range: %q", item)
				}
			}
			for v := lo; v <= hi; v += step {
				if i == 4 && v == 7 {
					c.fields[i][0] = true
				} else {
					c.fields[i][v] = true
				}
			}
		}
	}
	return c, nil
}

func cronNumber(s string, field int) (int, error) {
	if field == 3 {
		for i, name := range []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"} {
			if strings.EqualFold(s, name) {
				return i + 1, nil
			}
		}
	}
	if field == 4 {
		for i, name := range []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"} {
			if strings.EqualFold(s, name) {
				return i, nil
			}
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid cron value %q", s)
	}
	return v, nil
}

func (c *Cron) dayMatches(t time.Time) bool {
	dom, dow := c.fields[2][t.Day()], c.fields[4][int(t.Weekday())]
	if c.wild[2] || c.wild[4] {
		return dom && dow
	}
	return dom || dow
}

func (c *Cron) Matches(t time.Time) bool {
	t = t.UTC()
	return c.fields[0][t.Minute()] && c.fields[1][t.Hour()] && c.fields[3][int(t.Month())] && c.dayMatches(t)
}

// Next returns nil for expressions with no occurrence in the next eight years.
func (c *Cron) Next(after time.Time) *time.Time {
	t := after.UTC().Truncate(time.Minute).Add(time.Minute)
	limit := t.AddDate(8, 0, 0)
	for t.Before(limit) {
		if !c.fields[3][int(t.Month())] || !c.dayMatches(t) {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC)
			continue
		}
		if !c.fields[1][t.Hour()] {
			t = t.Truncate(time.Hour).Add(time.Hour)
			continue
		}
		if c.fields[0][t.Minute()] {
			return &t
		}
		t = t.Add(time.Minute)
	}
	return nil
}
