package domains

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateOnly represents a date without time component (YYYY-MM-DD), serialized matching C# DateOnly
type DateOnly struct {
	time.Time
}

func NewDateOnly(t time.Time) DateOnly {
	return DateOnly{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

func NowDateOnly() DateOnly {
	return NewDateOnly(time.Now())
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", d.Format("2006-01-02"))), nil
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		d.Time = t
		return nil
	}

	t, err = time.Parse(time.RFC3339Nano, s)
	if err == nil {
		d.Time = t
		return nil
	}

	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		d.Time = t
		return nil
	}

	t, err = time.Parse("2006-01-02T15:04:05.999999999", s)
	if err == nil {
		d.Time = t
		return nil
	}

	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		d.Time = t
		return nil
	}

	return err
}

func (d DateOnly) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		return d.UnmarshalJSON([]byte("\"" + v + "\""))
	case []byte:
		return d.UnmarshalJSON([]byte("\"" + string(v) + "\""))
	default:
		return fmt.Errorf("cannot scan type %T into DateOnly", value)
	}
}
