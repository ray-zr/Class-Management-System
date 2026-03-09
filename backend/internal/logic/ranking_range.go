package logic

import (
	"fmt"
	"time"
)

func rankingRange(now time.Time, month string, total bool, week int64) (start time.Time, end time.Time, isTotal bool, err error) {
	if total {
		return time.Time{}, time.Time{}, true, nil
	}
	if month == "" {
		month = now.Format("2006-01")
	}
	monthStart, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid month")
	}
	monthEnd := monthStart.AddDate(0, 1, 0)

	if week <= 0 {
		return monthStart, monthEnd, false, nil
	}
	weekStart := monthStart.AddDate(0, 0, int((week-1)*7))
	if !weekStart.Before(monthEnd) {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid week")
	}
	weekEnd := weekStart.AddDate(0, 0, 7)
	if weekEnd.After(monthEnd) {
		weekEnd = monthEnd
	}
	return weekStart, weekEnd, false, nil
}
