package timeutil

import (
	"testing"
	"time"
)

func TestStartOfMonth(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{"月中", time.Date(2026, 7, 29, 10, 6, 22, 0, time.Local), "2026-07-01"},
		{"恰好在月首", time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local), "2026-07-01"},
		{"月末", time.Date(2026, 12, 31, 23, 59, 59, 0, time.Local), "2026-12-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StartOfMonth(tt.in)
			if got.Format(time.DateOnly) != tt.want {
				t.Errorf("StartOfMonth(%s) = %s, want %s", tt.in.Format(time.DateTime), got.Format(time.DateOnly), tt.want)
			}
			if got.Location() != tt.in.Location() {
				t.Errorf("StartOfMonth 改变了时区: %v, want %v", got.Location(), tt.in.Location())
			}
		})
	}
}

func TestAddMonths(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		n    int
		want string
	}{
		{"往后一个月", time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local), 1, "2026-08-01"},
		{"往后三个月", time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local), 3, "2026-10-01"},
		{"跨年", time.Date(2026, 11, 1, 0, 0, 0, 0, time.Local), 3, "2027-02-01"},
		{"往前一个月", time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local), -1, "2026-06-01"},
		{"往前跨年", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), -1, "2025-12-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddMonths(tt.in, tt.n).Format(time.DateOnly); got != tt.want {
				t.Errorf("AddMonths(%s, %d) = %s, want %s", tt.in.Format(time.DateOnly), tt.n, got, tt.want)
			}
		})
	}
}

// 非月首起算会被 Go 归一化，所以调用方必须先 StartOfMonth——这条固化这个前提
func TestAddMonthsNormalizesFromNonMonthStart(t *testing.T) {
	got := AddMonths(time.Date(2026, 1, 31, 0, 0, 0, 0, time.Local), 1).Format(time.DateOnly)
	if got != "2026-03-03" {
		t.Errorf("AddMonths(2026-01-31, 1) = %s, want 2026-03-03", got)
	}
}
