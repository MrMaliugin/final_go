package api

import (
	"fmt"
	"time"
)

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	const layout = "20060102"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		return "", fmt.Errorf("неправильная дата: %v", err)
	}

	var nextDate time.Time
	switch repeat {
	case "":
		nextDate = date
	case "d":
		nextDate = date.AddDate(0, 0, 1)
	default:
		return "", fmt.Errorf("неподдерживаемое правило повторения: %s", repeat)
	}

	for nextDate.Before(now) {
		switch repeat {
		case "d":
			nextDate = nextDate.AddDate(0, 0, 1)
		default:
			return "", fmt.Errorf("неподдерживаемое правило повторения: %s", repeat)
		}
	}

	return nextDate.Format(layout), nil
}
