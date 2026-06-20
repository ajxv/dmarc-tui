package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"

	"dmarc-tui/internal/dmarc"
	"dmarc-tui/internal/ipinfo"
)

func newBareModel(records []dmarc.Record) model {
	d := recordDelegate{}
	l := list.New([]list.Item{}, d, 80, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	ti := textinput.New()
	ti.Prompt = ""
	return model{
		feedbacks: []*dmarc.Feedback{{Records: records}},
		fileIdx:   0,
		list:      l,
		textInput: ti,
		ipCache:   make(map[string]ipinfo.CacheEntry),
	}
}

func makeRecord(dkimResult, spfResult string) dmarc.Record {
	return dmarc.Record{
		Row: dmarc.Row{
			PolicyEvaluated: dmarc.PolicyEvaluated{DKIM: dkimResult, SPF: spfResult},
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
			rec: dmarc.Record{AuthResults: dmarc.AuthResults{
				SPF: []dmarc.SPFResult{{Domain: "spf.example.com"}},
			}},
			want: "spf.example.com",
		},
		{
			name: "spf domain that is a bare ip is skipped",
			rec: dmarc.Record{AuthResults: dmarc.AuthResults{
				SPF: []dmarc.SPFResult{{Domain: "203.0.113.5"}},
			}},
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

func TestApplyFilterSort(t *testing.T) {
	m := newBareModel([]dmarc.Record{
		{Row: dmarc.Row{SourceIP: "203.0.113.10", Count: 5, PolicyEvaluated: dmarc.PolicyEvaluated{DKIM: "pass", SPF: "pass"}}},
		{Row: dmarc.Row{SourceIP: "203.0.113.2", Count: 50, PolicyEvaluated: dmarc.PolicyEvaluated{DKIM: "fail", SPF: "fail"}}},
		{Row: dmarc.Row{SourceIP: "203.0.113.30", Count: 1, PolicyEvaluated: dmarc.PolicyEvaluated{DKIM: "pass", SPF: "fail"}}},
	})

	t.Run("filterPass keeps only dual-mechanism-passing records", func(t *testing.T) {
		m.filter = filterPass
		m.applyFilterSort()
		if len(m.filtered) != 1 {
			t.Fatalf("expected 1 passing record, got %d", len(m.filtered))
		}
		if dmarc.Classify(m.filtered[0]) != dmarc.ClassPass {
			t.Errorf("filterPass returned non-pass record: %+v", m.filtered[0])
		}
	})

	t.Run("filterReview keeps only single-mechanism-passing records", func(t *testing.T) {
		m.filter = filterReview
		m.applyFilterSort()
		if len(m.filtered) != 1 {
			t.Fatalf("expected 1 review record, got %d", len(m.filtered))
		}
		if dmarc.Classify(m.filtered[0]) != dmarc.ClassNeedsReview {
			t.Errorf("filterReview returned non-review record: %+v", m.filtered[0])
		}
	})

	t.Run("filterFail keeps only fully-failing records", func(t *testing.T) {
		m.filter = filterFail
		m.applyFilterSort()
		if len(m.filtered) != 1 || m.filtered[0].Row.SourceIP != "203.0.113.2" {
			t.Fatalf("expected exactly the failing record, got %+v", m.filtered)
		}
	})

	t.Run("sortVolume orders by descending count", func(t *testing.T) {
		m.filter = filterAll
		m.sort = sortVolume
		m.applyFilterSort()
		want := []int{50, 5, 1}
		for i, r := range m.filtered {
			if r.Row.Count != want[i] {
				t.Errorf("position %d: count = %d, want %d", i, r.Row.Count, want[i])
			}
		}
	})

	t.Run("sortIP orders numerically, not lexicographically, by parsed IP", func(t *testing.T) {
		m.sort = sortIP
		m.applyFilterSort()
		want := []string{"203.0.113.2", "203.0.113.10", "203.0.113.30"}
		for i, r := range m.filtered {
			if r.Row.SourceIP != want[i] {
				t.Errorf("position %d: ip = %s, want %s", i, r.Row.SourceIP, want[i])
			}
		}
	})

	t.Run("search query narrows results", func(t *testing.T) {
		m.filter = filterAll
		m.sort = sortVolume
		m.textInput.SetValue("203.0.113.2")
		m.applyFilterSort()
		if len(m.filtered) != 1 || m.filtered[0].Row.SourceIP != "203.0.113.2" {
			t.Fatalf("expected search to narrow to one record, got %+v", m.filtered)
		}
		m.textInput.SetValue("")
	})

	t.Run("cursor clamps to last item when filter shrinks the list", func(t *testing.T) {
		m.filter = filterAll
		m.applyFilterSort()
		m.list.Select(len(m.filtered) - 1)
		m.filter = filterFail
		m.applyFilterSort()
		want := max(0, len(m.filtered)-1)
		if m.list.Index() != want {
			t.Errorf("cursor = %d, want %d (clamped to last filtered index)", m.list.Index(), want)
		}
	})
}

func TestPassFailCounts(t *testing.T) {
	m := newBareModel([]dmarc.Record{
		makeRecord("pass", "pass"), // pass
		makeRecord("pass", "fail"), // needs review -> counted as pass
		makeRecord("fail", "pass"), // needs review -> counted as pass
		makeRecord("fail", "fail"), // fail
		makeRecord("fail", "fail"), // fail
	})
	pass, fail := m.passFailCounts()
	if pass != 3 || fail != 2 {
		t.Errorf("passFailCounts() = (%d, %d), want (3, 2)", pass, fail)
	}
}
