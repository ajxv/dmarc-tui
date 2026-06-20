package dmarc_test

import (
	"testing"

	"dmarc-tui/internal/dmarc"
)

func makeRecord(dkim, spf string) dmarc.Record {
	return dmarc.Record{
		Row: dmarc.Row{
			PolicyEvaluated: dmarc.PolicyEvaluated{DKIM: dkim, SPF: spf},
		},
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name         string
		dkim, spf    string
		wantClassify dmarc.Classification
	}{
		{"both pass", "pass", "pass", dmarc.ClassPass},
		{"dkim only", "pass", "fail", dmarc.ClassNeedsReview},
		{"spf only", "fail", "pass", dmarc.ClassNeedsReview},
		{"neither", "fail", "fail", dmarc.ClassFail},
		{"both missing", "", "", dmarc.ClassFail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dmarc.Classify(makeRecord(tt.dkim, tt.spf)); got != tt.wantClassify {
				t.Errorf("Classify(dkim=%q, spf=%q) = %v, want %v", tt.dkim, tt.spf, got, tt.wantClassify)
			}
		})
	}
}

func TestIsAligned(t *testing.T) {
	tests := []struct {
		name                         string
		authDomain, headerFrom, mode string
		want                         bool
	}{
		{"exact match", "example.com", "example.com", "r", true},
		{"exact match strict", "example.com", "example.com", "s", true},
		{"relaxed org domain match", "mail.example.com", "news.example.com", "r", true},
		{"relaxed org domain match default mode", "mail.example.com", "news.example.com", "", true},
		{"strict mismatch", "mail.example.com", "news.example.com", "s", false},
		{"relaxed org domain mismatch", "example.com", "other.com", "r", false},
		{"multi-label tld relaxed match", "mail.example.co.uk", "shop.example.co.uk", "r", true},
		{"multi-label tld relaxed mismatch", "mail.example.co.uk", "mail.other.co.uk", "r", false},
		{"empty auth domain", "", "example.com", "r", false},
		{"empty header from", "example.com", "", "r", false},
		{"case insensitive", "MAIL.EXAMPLE.COM", "news.example.com", "r", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dmarc.IsAligned(tt.authDomain, tt.headerFrom, tt.mode); got != tt.want {
				t.Errorf("IsAligned(%q, %q, %q) = %v, want %v", tt.authDomain, tt.headerFrom, tt.mode, got, tt.want)
			}
		})
	}
}

func TestAuthDomain(t *testing.T) {
	tests := []struct {
		name string
		rec  dmarc.Record
		want string
	}{
		{
			name: "dkim domain present",
			rec: dmarc.Record{AuthResults: dmarc.AuthResults{
				DKIM: []dmarc.DKIMResult{{Domain: "example.com"}},
				SPF:  []dmarc.SPFResult{{Domain: "spf.example.com"}},
			}},
			want: "example.com",
		},
		{
			name: "falls back to spf domain when no dkim",
			rec:  dmarc.Record{AuthResults: dmarc.AuthResults{SPF: []dmarc.SPFResult{{Domain: "spf.example.com"}}}},
			want: "spf.example.com",
		},
		{
			name: "spf domain that is a bare ip is skipped",
			rec:  dmarc.Record{AuthResults: dmarc.AuthResults{SPF: []dmarc.SPFResult{{Domain: "203.0.113.5"}}}},
			want: "",
		},
		{name: "no dkim or spf domain", rec: dmarc.Record{}, want: ""},
		{
			name: "dkim entry present but domain empty falls back to spf",
			rec: dmarc.Record{AuthResults: dmarc.AuthResults{
				DKIM: []dmarc.DKIMResult{{Domain: ""}},
				SPF:  []dmarc.SPFResult{{Domain: "spf.example.com"}},
			}},
			want: "spf.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dmarc.AuthDomain(tt.rec); got != tt.want {
				t.Errorf("AuthDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecordMatches(t *testing.T) {
	rec := dmarc.Record{
		Row:         dmarc.Row{SourceIP: "203.0.113.5"},
		Identifiers: dmarc.Identifiers{HeaderFrom: "news.example.com"},
		AuthResults: dmarc.AuthResults{DKIM: []dmarc.DKIMResult{{Domain: "mail.example.com"}}},
	}
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{"matches source ip", "203.0.113", true},
		{"matches authenticated domain", "mail.example", true},
		{"matches header from", "news.example.com", true},
		{"case insensitive haystack, lower-cased query", "mail.example", true},
		{"no match", "nope", false},
		{"empty query matches nothing special but is contained", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dmarc.RecordMatches(rec, tt.query); got != tt.want {
				t.Errorf("RecordMatches(q=%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
