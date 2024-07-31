package sqlite

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// NullTime represents a wrapper around time.Time, providing support for
// storing and reading time values formatted using time.RFC3339 to and
// from SQLite databases. NULL values are converted to zero-value times.
type NullTime time.Time

// Scan implements the sql.Scanner interface.
func (n *NullTime) Scan(src any) error {
	if src == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	}

	str, ok := src.(string)
	if ok {
		var err error

		*(*time.Time)(n), err = time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("NullTime: cannot scan to time.Time: %T", src)
		}

		return nil
	}

	return fmt.Errorf("NullTime: cannot scan to time.Time: %T", src)
}

// Value implements the driver.Valuer interface.
func (n *NullTime) Value() (driver.Value, error) {
	if n == nil || (*time.Time)(n).IsZero() {
		return nil, nil
	}

	return (*time.Time)(n).UTC().Format(time.RFC3339), nil
}
