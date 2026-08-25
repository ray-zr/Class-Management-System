package logic

func normalizeRollcallCount(count int64) int64 {
	if count < 1 {
		return 1
	}
	if count > 50 {
		return 50
	}
	return count
}
