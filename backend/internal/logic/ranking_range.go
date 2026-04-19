package logic

import (
	"fmt"
	"time"
)

func rankingRange(now time.Time, month, startDate, endDate string, total bool) (start time.Time, end time.Time, isTotal bool, err error) {
	if total {
		return time.Time{}, time.Time{}, true, nil
	}
	if (startDate == "") != (endDate == "") {
		return time.Time{}, time.Time{}, false, fmt.Errorf("startDate and endDate must be provided together")
	}
	if startDate != "" && endDate != "" {
		start, err = time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, false, fmt.Errorf("invalid startDate")
		}
		endDay, parseErr := time.ParseInLocation("2006-01-02", endDate, time.Local)
		if parseErr != nil {
			return time.Time{}, time.Time{}, false, fmt.Errorf("invalid endDate")
		}
		if endDay.Before(start) {
			return time.Time{}, time.Time{}, false, fmt.Errorf("endDate must be on or after startDate")
		}
		return start, endDay.AddDate(0, 0, 1), false, nil
	}
	if month == "" {
		month = now.Format("2006-01")
	}
	monthStart, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid month")
	}
	return monthStart, monthStart.AddDate(0, 1, 0), false, nil
}
