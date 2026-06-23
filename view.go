package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"dmarc-tui/internal/dmarc"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── bubbles/list integration ──────────────────────────────────────────────────

// recordItem wraps a Record so it satisfies list.Item.
type recordItem struct {
	record dmarc.Record
}

func (r recordItem) FilterValue() string { return r.record.Row.SourceIP }

// recordDelegate renders each record as a styled table row. Column widths and
// totalMsgs are set by the model whenever window size or file changes.
type recordDelegate struct {
	wSrc, wProv int
	tableInnerW int
	totalMsgs   int
}

func (d recordDelegate) Height() int                             { return 1 }
func (d recordDelegate) Spacing() int                            { return 1 }
func (d recordDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d recordDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ri, ok := item.(recordItem)
	if !ok {
		return
	}
	r := ri.record
	isSel := index == m.Index()

	dk := r.Row.PolicyEvaluated.DKIM
	sp := r.Row.PolicyEvaluated.SPF
	dmarcOk := dk == "pass" || sp == "pass"

	ipTrunc := trunc(r.Row.SourceIP, d.wSrc)
	var ipStyled string
	if isSel {
		ipStyled = styleWhiteBold.Render(ipTrunc)
	} else {
		ipStyled = styleWhite.Render(ipTrunc)
	}
	srcCell := padRight(ipStyled, d.wSrc)
	provCell := padRight(domainStyled(trunc(dmarc.AuthDomain(r), d.wProv), isSel), d.wProv)
	authCell := padRight(authInline(dmarcOk, dk, sp, isSel), colWidthAuth)

	vol := volumeBar(r.Row.Count, d.totalMsgs, volBarWidth)
	countStyle := styleWhite
	if isSel {
		countStyle = styleWhiteBold
	}
	numStr := formatNum(r.Row.Count)
	numW := colWidthVol - volBarWidth - 1
	pad := max(0, numW-len(numStr))
	volCell := padRight(vol+" "+strings.Repeat(" ", pad)+countStyle.Render(numStr), colWidthVol)
	clsCell := classBadge(dmarc.Classify(r))

	row := srcCell + "  " + provCell + "  " + authCell + "  " + volCell + "  " + clsCell
	cursor := "  "
	if isSel {
		cursor = styleAccentBold.Render("›") + " "
	}
	// The cursor stays on the app background; only the row content carries the
	// selection tint, so the "›" marker isn't boxed in a highlight.
	inner := padRight(row, d.tableInnerW-1-2)
	if isSel {
		inner = lipgloss.NewStyle().Background(colSelectionBg).Render(inner)
	}
	fmt.Fprint(w, cursor+inner)
}

// ── Top-level View ────────────────────────────────────────────────────────────

func (m model) View() string {
	if !m.ready {
		return "\n  Loading…"
	}
	// Hard minimum: below minTermWidth×minTermHeight the table columns overflow
	// and text truncates badly. Show a single clear message instead of a broken
	// layout; the user just needs to make the window bigger.
	if m.width < minTermWidth || m.height < minTermHeight {
		msg := fmt.Sprintf("Terminal too small (%d×%d) — resize to at least %d×%d",
			m.width, m.height, minTermWidth, minTermHeight)
		return fillLine("  "+styleRed.Render(trunc(msg, max(m.width-4, 0))),
			max(m.width, 1), colAppBg)
	}
	var body string
	if m.helpVisible {
		body = m.renderHelpView()
	} else if m.inspecting {
		body = m.renderDetailView()
	} else {
		blank := fillLine("", m.contentWidth(), colAppBg)
		body = m.renderTopBar() + "\n" +
			blank + "\n" +
			m.renderStatCards() + "\n" +
			blank + "\n" +
			m.renderFilterBar() + "\n" +
			blank + "\n" +
			m.renderContent() +
			m.renderStatusBar()
	}
	return body
}

// ── Top Bar ───────────────────────────────────────────────────────────────────

func (m model) renderTopBar() string {
	fb := m.feedbacks[m.fileIdx]
	pol := fb.PolicyPublished

	brand := lipgloss.NewStyle().Foreground(colBrand).Bold(true).Render("DMARC-TUI")
	sep := styleMuted.Render(" | ")
	reports := styleMuted.Render("Reports")
	arrow := styleMuted.Render(" → ")
	domain := styleWhite.Render(pol.Domain)
	sources := styleWhiteBold.Render("Sources")

	// orgBadge names the reporting organization (e.g. "Google", "Mimecast") —
	// the detail that actually answers "which report am I looking at?", since
	// file position alone ("1 / 8") doesn't say who sent it.
	orgBadge := lipgloss.NewStyle().
		Foreground(colCyan).
		Padding(0, 1).
		Render(trunc(fb.ReportMetadata.OrgName, 24))

	left := " " + brand + sep + orgBadge + sep + reports + arrow + domain + arrow + sources

	navL, navR := styleDimMuted.Render("‹"), styleDimMuted.Render("›")
	if m.fileIdx > 0 {
		navL = styleMuted.Render("‹")
	}
	if m.fileIdx < len(m.feedbacks)-1 {
		navR = styleMuted.Render("›")
	}
	repStr := fmt.Sprintf("%d / %d", m.fileIdx+1, len(m.feedbacks))
	right := styleMuted.Render("report ") + navL + " " + styleWhite.Render(repStr) + " " + navR + " "

	return fillLine(splitLine(left, right, m.contentWidth()), m.contentWidth(), colAppBg)
}

// ── Stat Cards ────────────────────────────────────────────────────────────────

func (m model) renderStatCards() string {
	fb := m.feedbacks[m.fileIdx]
	meta := fb.ReportMetadata
	begin := time.Unix(meta.DateRange.Begin, 0).UTC()
	end := time.Unix(meta.DateRange.End, 0).UTC()

	var total, dmarcPass, spfPass, dkimPass int
	for _, r := range fb.Records {
		n := r.Row.Count
		total += n
		dk := r.Row.PolicyEvaluated.DKIM
		sp := r.Row.PolicyEvaluated.SPF
		if dk == "pass" || sp == "pass" {
			dmarcPass += n
		}
		if sp == "pass" {
			spfPass += n
		}
		if dk == "pass" {
			dkimPass += n
		}
	}
	passCount, failCount := m.passFailCounts()
	totalSrc := len(fb.Records)

	W := m.contentWidth()
	areaW := W - 4
	barsW := max(areaW*barsWidthPct/100, 30)
	cardsW := areaW - barsW
	eachW := cardsW / 4
	lastW := cardsW - eachW*3

	var rangeStr string
	switch {
	case begin.Format("2006-01-02") == end.Format("2006-01-02"):
		rangeStr = begin.Format("Jan 02, 2006")
	case begin.Year() == end.Year():
		rangeStr = begin.Format("Jan 02") + " – " + end.Format("Jan 02, 2006")
	default:
		rangeStr = begin.Format("Jan 02, 2006") + " – " + end.Format("Jan 02, 2006")
	}
	timeStr := begin.Format("15:04") + " – " + end.Format("15:04") + " UTC"

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colDimMuted)

	card := func(label, value, sub string, totalW int) string {
		return cardStyle.Width(totalW - 2).
			PaddingLeft(1).
			Render(lipgloss.JoinVertical(lipgloss.Left,
				styleMuted.Render(label),
				value,
				sub,
			))
	}

	barsInner := barsW - 2
	barW := max(barsInner-16, 8)
	barsPanel := cardStyle.Width(barsInner).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			barRow("DMARC", dmarcPass, total, barW),
			barRow("SPF", spfPass, total, barW),
			barRow("DKIM", dkimPass, total, barW),
		))

	joined := lipgloss.JoinHorizontal(lipgloss.Top,
		card("PERIOD", styleWhiteBold.Render(trunc(rangeStr, eachW-4)), styleMuted.Render(timeStr), eachW),
		card("MESSAGES", styleWhiteBold.Render(formatNum(total))+styleMuted.Render(" total"), styleMuted.Render(fmt.Sprintf("%d sources", len(fb.Records))), eachW),
		card("PASS", styleGreenBold.Render(fmt.Sprintf("%d", passCount))+styleMuted.Render(fmt.Sprintf(" / %d", totalSrc)), styleMuted.Render("sources"), eachW),
		card("FAIL", styleRedBold.Render(fmt.Sprintf("%d", failCount))+styleMuted.Render(fmt.Sprintf(" / %d", totalSrc)), styleMuted.Render("sources"), lastW),
		barsPanel,
	)

	var sb strings.Builder
	for i, line := range strings.Split(joined, "\n") {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(fillLine("  "+line, W, colAppBg))
	}
	return sb.String()
}

func barRow(label string, val, total, barW int) string {
	p := pct(val, total)
	bar := lineBar(val, total, barW)
	pctStr := passRateBadge(p)
	return " " + padRight(styleMuted.Render(label), 5) + " " + bar + " " + pctStr
}

// ── Filter Bar ────────────────────────────────────────────────────────────────

func (m model) renderFilterBar() string {
	if m.searching {
		return m.renderSearchBar()
	}

	filterLabel := styleMuted.Render("FILTER")

	filterNames := []string{"All", "Passing", "Review", "Failing"}
	filterDots := []string{"", "● ", "● ", "● "}
	filterDotStyles := []*lipgloss.Style{nil, &styleGreen, &styleYellow, &styleRed}

	var pills []string
	for i, name := range filterNames {
		dot := ""
		if filterDots[i] != "" {
			dot = filterDotStyles[i].Render(filterDots[i])
		}
		if filterMode(i) == m.filter {
			pills = append(pills, dot+styleWhiteBold.Underline(true).Render(name))
		} else {
			pills = append(pills, dot+styleMuted.Render(name))
		}
	}

	left := "  " + filterLabel + "  " + strings.Join(pills, "  ")
	if v := m.textInput.Value(); v != "" {
		badge := styleMuted.Render("⌕ ") + styleWhiteBold.Render(v) +
			styleDimMuted.Render("  (/ edit)")
		left += "    " + badge
	}

	sortLabel := styleMuted.Render("SORT")
	sortNames := []string{"Volume", "IP"}
	sortArrows := []string{"↓", "↑"}
	var sortParts []string
	for i, name := range sortNames {
		if sortMode(i) == m.sort {
			sortParts = append(sortParts, styleWhiteBold.Render(name+sortArrows[i]))
		} else {
			sortParts = append(sortParts, styleMuted.Render(name))
		}
	}
	right := sortLabel + "  " + strings.Join(sortParts, "  ") + "  "

	return fillLine(splitLine(left, right, m.contentWidth()), m.contentWidth(), colAppBg)
}

// renderSearchBar replaces the filter bar while a search is being typed —
// a live text prompt with the matching-record count and key hints.
func (m model) renderSearchBar() string {
	label := styleMuted.Render("SEARCH")
	input := m.textInput.View()
	count := styleMuted.Render(fmt.Sprintf("%d match", len(m.filtered)))
	if len(m.filtered) != 1 {
		count = styleMuted.Render(fmt.Sprintf("%d matches", len(m.filtered)))
	}
	left := "  " + label + "   " + input + "   " + count

	key := func(k, desc string) string {
		return styleWhiteBold.Render(k) + " " + styleMuted.Render(desc)
	}
	right := key("enter", "apply") + "   " + key("esc", "cancel") + "  "

	return fillLine(splitLine(left, right, m.contentWidth()), m.contentWidth(), colAppBg)
}

// ── Status Bar ────────────────────────────────────────────────────────────────

func (m model) renderStatusBar() string {
	h := m.help
	h.Width = m.contentWidth() - 4
	return fillLine("  "+h.View(listKeys), m.contentWidth(), colAppBg)
}

// ── Content (records table) ───────────────────────────────────────────────────

func (m model) renderContent() string {
	return m.renderRecordsTable(max(m.contentWidth(), 60))
}

func (m model) renderRecordsTable(W int) string {
	tableInnerW, wSrc, wProv, listH := m.listGeometry()
	totalSrc := len(m.feedbacks[m.fileIdx].Records)

	var sb strings.Builder
	wl := func(s string) {
		sb.WriteString(fillLine(s, W, colAppBg))
		sb.WriteByte('\n')
	}
	bl := styleBorder.Render
	bRight := bl("│")

	// Top border
	srcCount := fmt.Sprintf("SOURCES · %d", totalSrc)
	if len(m.filtered) != totalSrc {
		srcCount = fmt.Sprintf("SOURCES · %d / %d", len(m.filtered), totalSrc)
	}
	labelStr := styleWhiteBold.Render(srcCount)
	labelVisW := lipgloss.Width(labelStr)
	dashW := max(tableInnerW-labelVisW-4, 0)
	wl("  " + bl("╭─ ") + labelStr + bl(" "+strings.Repeat("─", dashW)+"╮"))

	// Column headers
	hdr := padRight(styleMuted.Render("SOURCE IP"), wSrc) + "  " +
		padRight(styleMuted.Render("DOMAIN"), wProv) + "  " +
		padRight(styleMuted.Render("AUTHENTICATION"), colWidthAuth) + "  " +
		padRight(styleMuted.Render("VOLUME"), colWidthVol) + "  " +
		styleMuted.Render("CLASS")
	wl("  " + bl("│") + padRight("  "+hdr, tableInnerW-1) + bRight)
	wl("  " + bl("│") + strings.Repeat(" ", tableInnerW-1) + bRight)

	// Records — list.View() handles scrolling, selection, and spacing.
	if len(m.filtered) == 0 {
		empty := padRight("    "+styleMuted.Render("(no records match current filter)"), tableInnerW-1)
		wl("  " + bl("│") + empty + bRight)
		for i := 1; i < listH; i++ {
			wl("  " + bl("│") + strings.Repeat(" ", tableInnerW-1) + bRight)
		}
	} else {
		lines := strings.Split(strings.TrimRight(m.list.View(), "\n"), "\n")
		for _, line := range lines {
			wl("  " + bl("│") + padRight(line, tableInnerW-1) + bRight)
		}
		for i := len(lines); i < listH; i++ {
			wl("  " + bl("│") + strings.Repeat(" ", tableInnerW-1) + bRight)
		}
	}

	wl("  " + bl("╰"+strings.Repeat("─", tableInnerW-1)+"╯"))
	wl("")
	return sb.String()
}
