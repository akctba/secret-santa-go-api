package database

import (
	"fmt"
	"time"
)

var sqliteTimeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05-07:00",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseDBTime(value any) (time.Time, error) {
	switch typedValue := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return typedValue, nil
	case string:
		return parseDBTimeString(typedValue)
	case []byte:
		return parseDBTimeString(string(typedValue))
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type %T", value)
	}
}

func parseDBTimeString(value string) (time.Time, error) {
	for _, layout := range sqliteTimeFormats {
		parsedTime, err := time.Parse(layout, value)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time value %q", value)
}
