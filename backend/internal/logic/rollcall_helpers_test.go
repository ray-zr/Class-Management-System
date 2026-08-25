package logic

import "testing"

func TestNormalizeRollcallCount(t *testing.T) {
	tests := []struct {
		input int64
		want  int64
	}{
		{input: -1, want: 1},
		{input: 0, want: 1},
		{input: 20, want: 20},
		{input: 51, want: 50},
	}
	for _, tt := range tests {
		if got := normalizeRollcallCount(tt.input); got != tt.want {
			t.Errorf("normalizeRollcallCount(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
