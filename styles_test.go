package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestFormatNum(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{7, "7"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{1234567, "1,234,567"},
	}
	for _, tt := range tests {
		if got := formatNum(tt.in); got != tt.want {
			t.Errorf("formatNum(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPct(t *testing.T) {
	const epsilon = 1e-9
	tests := []struct {
		a, b int
		want float64
	}{
		{50, 100, 50},
		{1, 3, 33.333333333333336},
		{0, 0, 0},
		{5, 0, 0},
	}
	for _, tt := range tests {
		got := pct(tt.a, tt.b)
		diff := got - tt.want
		if diff < -epsilon || diff > epsilon {
			t.Errorf("pct(%d, %d) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTrunc(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"short", 10, "short"},
		{"exactlyten", 10, "exactlyten"},
		{"this is too long", 8, "this is…"},
		{"日本語のテスト", 4, "日本語…"},
	}
	for _, tt := range tests {
		if got := trunc(tt.in, tt.n); got != tt.want {
			t.Errorf("trunc(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		in    string
		width int
		want  string
	}{
		{"abc", 5, "abc  "},
		{"abc", 3, "abc"},
		{"abc", 2, "abc"},
	}
	for _, tt := range tests {
		if got := padRight(tt.in, tt.width); got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
		}
	}
}

func TestSplitLine(t *testing.T) {
	tests := []struct {
		left, right string
		W           int
		wantLen     int
	}{
		{"DMARC-TUI", "report ‹ 1/1 ›", 40, 40},
		// gap floors at 1 when combined width already meets or exceeds W
		{"left", "right", 5, 10}, // "left" + 1 space + "right" = 10
		{"", "", 10, 10},
	}
	for _, tt := range tests {
		got := splitLine(tt.left, tt.right, tt.W)
		if lipgloss.Width(got) != tt.wantLen {
			t.Errorf("splitLine(%q, %q, %d): visual width = %d, want %d",
				tt.left, tt.right, tt.W, lipgloss.Width(got), tt.wantLen)
		}
		// right content must appear at the end
		if !strings.HasSuffix(got, tt.right) {
			t.Errorf("splitLine result %q does not end with right %q", got, tt.right)
		}
	}
}
