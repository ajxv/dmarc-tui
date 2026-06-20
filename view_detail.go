package main

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"dmarc-tui/internal/dmarc"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// ── Detail View ───────────────────────────────────────────────────────────────

func (m model) renderDetailView() string {
	if m.list.Index() >= len(m.filtered) {
		return ""
	}
	r := m.filtered[m.list.Index()]
	W := m.contentWidth()
	return fillLine(m.renderDetailHeader(r), W, colAppBg) + "\n" +
		fillLine("", W, colAppBg) + "\n" +
		m.detailVP.View() + "\n" +
		m.renderDetailStatusBar()
}

// renderDetailBody produces the scrollable content set into detailVP: hero +
// tab bar + tab content. Each line is pre-filled to W so the viewport
// background stays consistent when scrolling.
func (m model) renderDetailBody(r dmarc.Record, W int) string {
	var sb strings.Builder
	fill := func(s string) { sb.WriteString(fillLine(s, W, colAppBg)); sb.WriteByte('\n') }

	m.renderDetailHeroInto(&sb, r, W)
	sb.WriteString(fillLine(m.renderDetailTabBar(), W, colAppBg))
	sb.WriteByte('\n')
	fill("")

	var content string
	switch m.detailTab {
	case tabEvaluation:
		content = m.renderDetailEvaluation(r, W)
	case tabIdentifiers:
		content = m.renderDetailIdentifiers(r)
	case tabPolicy:
		content = m.renderDetailPolicy()
	case tabRawXML:
		content = m.renderDetailRawXML(r)
	case tabIPInfo:
		content = m.renderDetailIPInfo(r.Row.SourceIP)
	}

	bordered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colDimMuted).
		Width(W-4).
		Padding(1, 0).
		Render(content)
	for line := range strings.SplitSeq(bordered, "\n") {
		fill("  " + line)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (m model) renderDetailHeader(r dmarc.Record) string {
	fb := m.feedbacks[m.fileIdx]
	pol := fb.PolicyPublished

	brand := lipgloss.NewStyle().Foreground(colBrand).Bold(true).Render("DMARC-TUI")
	sep := styleMuted.Render(" | ")
	arrow := styleMuted.Render(" → ")
	domain := styleWhite.Render(pol.Domain)
	sources := styleMuted.Render("Sources")
	ip := styleWhiteBold.Render(r.Row.SourceIP)
	left := " " + brand + sep + domain + arrow + sources + arrow + ip

	navL, navR := styleDimMuted.Render("‹"), styleDimMuted.Render("›")
	if m.list.Index() > 0 {
		navL = styleMuted.Render("‹")
	}
	if m.list.Index() < len(m.filtered)-1 {
		navR = styleMuted.Render("›")
	}
	recStr := fmt.Sprintf("%d / %d", m.list.Index()+1, len(m.filtered))
	right := styleMuted.Render("record ") + navL + " " + styleWhite.Render(recStr) + " " + navR + " "

	return fillLine(splitLine(left, right, m.contentWidth()), m.contentWidth(), colAppBg)
}

func (m model) renderDetailHeroInto(sb *strings.Builder, r dmarc.Record, W int) {
	fb := m.feedbacks[m.fileIdx]
	meta := fb.ReportMetadata
	begin := time.Unix(meta.DateRange.Begin, 0).UTC()

	dk := r.Row.PolicyEvaluated.DKIM
	sp := r.Row.PolicyEvaluated.SPF
	dmarcOk := dk == "pass" || sp == "pass"

	fill := func(s string) { sb.WriteString(fillLine(s, W, colAppBg)); sb.WriteByte('\n') }

	var iconClr lipgloss.Color
	var iconChar string
	if dmarcOk {
		iconClr, iconChar = colGreen, "✓"
	} else {
		iconClr, iconChar = colRed, "✗"
	}

	iconBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(iconClr).
		Render(lipgloss.NewStyle().Foreground(iconClr).Bold(true).Render("  " + iconChar + "  "))
	iconLines := strings.Split(iconBox, "\n")
	iconTop := "  " + iconLines[0] + "  "
	iconMid := "  " + iconLines[1] + "  "
	iconBot := "  " + iconLines[2] + "  "

	const labelW = 16 // fixed label column width — "Sending domain" is the longest

	// Auth column: stacked vertically, one check per hero row, right-aligned.
	const authLabelW = 6 // "DMARC " covers all three labels
	authEntry := func(label string, ok bool) string {
		return padRight(styleMuted.Render(label), authLabelW) + statusPassFail(ok)
	}
	authLine1 := authEntry("DMARC", dmarcOk)
	authLine2 := authEntry("DKIM", dk == "pass")
	authLine3 := authEntry("SPF", sp == "pass")
	authW := lipgloss.Width(authLine1)

	msgWord := "message"
	if r.Row.Count != 1 {
		msgWord = "messages"
	}

	// authCol is the fixed column where the DMARC/DKIM/SPF status block starts.
	// A stable column gives consistent alignment across terminal widths — the
	// auth statuses don't jump around when resizing. On very narrow terminals
	// it falls back to 2 chars before the right edge.
	const authColStart = 100
	authCol := min(authColStart, W-authW-2)
	if authCol < 1 {
		authCol = 1
	}

	// Line 1 (iconTop): Source IP + DMARC at fixed column
	l1 := iconTop + padRight(styleMuted.Render("Source IP"), labelW) + styleWhiteBold.Render(r.Row.SourceIP)
	fill(l1 + strings.Repeat(" ", max(authCol-lipgloss.Width(l1), 1)) + authLine1)

	// Line 2 (iconMid): Sending domain + DKIM at fixed column
	authDom := styleMuted.Render("—")
	if dom := dmarc.AuthDomain(r); dom != "" {
		authDom = styleCyan.Render(dom)
	}
	l2 := iconMid + padRight(styleMuted.Render("Sending domain"), labelW) + authDom
	fill(l2 + strings.Repeat(" ", max(authCol-lipgloss.Width(l2), 1)) + authLine2)

	// Line 3 (iconBot): Email From + count + date + SPF at fixed column.
	l3 := iconBot + padRight(styleMuted.Render("Email From"), labelW) +
		styleCyan.Render(r.Identifiers.HeaderFrom) +
		styleMuted.Render("   ·   ") +
		styleWhite.Render(formatNum(r.Row.Count)) + " " + styleMuted.Render(msgWord) +
		styleMuted.Render("   ·   ") +
		styleMuted.Render(begin.Format("Jan 02, 2006"))
	fill(l3 + strings.Repeat(" ", max(authCol-lipgloss.Width(l3), 1)) + authLine3)

	fill("")
}

// renderDetailTabBar draws the section switcher as a segmented strip.
// The active section gets the "[ bracket ]" treatment; thin dividers separate
// each segment so they read as discrete units.
func (m model) renderDetailTabBar() string {
	tabs := []string{"Evaluation", "Identifiers", "Policy", "Raw XML", "IPInfo"}
	div := styleDimMuted.Render(" │ ")

	var parts []string
	for i, name := range tabs {
		label := fmt.Sprintf("%d %s", i+1, name)
		if detailTabID(i) == m.detailTab {
			parts = append(parts, outlineBadge(label, colAccent, 0))
		} else {
			parts = append(parts, styleMuted.Render(label))
		}
	}
	left := "  " + strings.Join(parts, div)

	key := func(k string) string { return styleWhiteBold.Render(k) }
	hint := styleMuted.Render("section  ") +
		key("tab") + styleMuted.Render(" · ") + key("←→") + styleMuted.Render(" · ") + key("1-5") + "  "

	return splitLine(left, hint, m.contentWidth())
}

// renderDetailIPInfo renders the IPInfo tab. It shows a spinner while the
// fetch is in flight, a formatted field table on success, and a specific error
// message on failure — including bogon / private addresses.
func (m model) renderDetailIPInfo(ip string) string {
	entry, ok := m.ipCache[ip]
	if !ok || !entry.Done {
		return "\n  " + styleMuted.Render("Fetching IPInfo for ") + styleWhite.Render(ip) + styleMuted.Render(" …")
	}

	if entry.Err != "" {
		return "\n  " + styleRed.Render("✗") + "  " + styleMuted.Render("Could not fetch IPInfo") +
			"\n\n  " + styleMuted.Render(entry.Err) +
			"\n\n  " + styleMuted.Render("press ") + styleWhiteBold.Render("r") + styleMuted.Render(" to retry")
	}

	d := entry.Info
	if d.Bogon {
		return "\n  " + styleMuted.Render(ip+" is a private or reserved address — no public data available")
	}

	type field struct{ label, value string }
	fields := []field{
		{"IP", d.IP},
		{"Hostname", d.Hostname},
		{"City", d.City},
		{"Region", d.Region},
		{"Country", d.Country},
		{"Coordinates", d.Loc},
		{"Organization", d.Org},
		{"Postal", d.Postal},
		{"Timezone", d.Timezone},
	}

	const labelW = 16
	var sb strings.Builder
	sb.WriteByte('\n')
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		sb.WriteString("  " + padRight(styleMuted.Render(f.label), labelW) + styleWhite.Render(f.value) + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (m model) renderDetailStatusBar() string {
	h := m.help
	h.Width = m.contentWidth() - 4
	return fillLine("  "+h.View(detailKeys), m.contentWidth(), colAppBg)
}

// ── Detail Tabs ───────────────────────────────────────────────────────────────

func (m model) renderDetailEvaluation(r dmarc.Record, W int) string {
	dk := r.Row.PolicyEvaluated.DKIM
	sp := r.Row.PolicyEvaluated.SPF
	dmarcOk := dk == "pass" || sp == "pass"

	overall := "fail"
	if dmarcOk {
		overall = "pass"
	}

	cls := dmarc.Classify(r)
	clsLine := classBadge(cls)
	if cls == dmarc.ClassNeedsReview {
		clsLine += styleMuted.Render("   DMARC passed on one mechanism only — the other didn't align")
	}

	var dkimResult, dkimDomain, dkimSel string
	if len(r.AuthResults.DKIM) > 0 {
		d := r.AuthResults.DKIM[0]
		dkimResult, dkimDomain, dkimSel = d.Result, d.Domain, d.Selector
	} else {
		dkimResult = "none"
	}

	var spfResult, spfDomain string
	if len(r.AuthResults.SPF) > 0 {
		s := r.AuthResults.SPF[0]
		spfResult, spfDomain = s.Result, s.Domain
	} else {
		spfResult = "none"
	}

	labelStyle := styleMuted.PaddingRight(3)
	cellStyle := lipgloss.NewStyle()
	colStyle := func(_, col int) lipgloss.Style {
		if col == 0 {
			return labelStyle
		}
		return cellStyle
	}

	// RESULT section: overall verdict, what the receiver did, and the
	// finer-grained dmarc.Classification — all three answer "what was the outcome?"
	// and belong together, distinct from the mechanism details below.
	resultTable := table.New().
		BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderHeader(false).
		StyleFunc(colStyle).
		Row("DMARC", resultStr(overall)).
		Row("Disposition", dispositionGloss(r.Row.PolicyEvaluated.Disposition)).
		Row("Classification", clsLine)

	var sb strings.Builder
	addLine := func(s string) { sb.WriteString(s); sb.WriteByte('\n') }

	addLine("  " + styleMuted.Render("RESULT"))
	addLine("")
	renderTableInto(&sb, resultTable)
	addLine("")
	addLine("")
	addLine("  " + styleMuted.Render("MECHANISMS"))
	addLine("")
	// MECHANISMS section: per-protocol alignment verdicts with domain and
	// selector details. mechDetail explains what the result means in context;
	// renderMechRow wraps long detail text with continuation lines aligned
	// under the detail column rather than back to col 0.
	renderMechRow(&sb, "DKIM", resultStr(dk),
		mechDetail(dkimResult, dkimDomain, dkimSel, r.Identifiers.HeaderFrom, dk == "pass"), W)
	renderMechRow(&sb, "SPF", resultStr(sp),
		mechDetail(spfResult, spfDomain, "", r.Identifiers.HeaderFrom, sp == "pass"), W)

	// The report's own stated reason for the policy outcome (policy_evaluated
	// /reason: a type plus an optional free-text comment) — surfaced only
	// when the report actually includes one, since most records carry none
	// and we don't invent an explanation where the data has none.
	if reasons := r.Row.PolicyEvaluated.Reasons; len(reasons) > 0 {
		addLine("")
		addLine("")
		addLine("  " + styleMuted.Render("REPORTED REASON"))
		addLine("")
		reasonTable := table.New().
			BorderTop(false).BorderBottom(false).
			BorderLeft(false).BorderRight(false).
			BorderColumn(false).BorderHeader(false).
			StyleFunc(colStyle)
		for i, rs := range reasons {
			label := "Reason"
			if len(reasons) > 1 {
				label = fmt.Sprintf("Reason %d", i+1)
			}
			typ := rs.Type
			if typ == "" {
				typ = "—"
			}
			val := styleYellow.Render(typ)
			if rs.Comment != "" {
				val += styleMuted.Render("   " + rs.Comment)
			}
			reasonTable.Row(label, val)
		}
		renderTableInto(&sb, reasonTable)
	}

	return strings.TrimRight(sb.String(), "\n")
}

func (m model) renderDetailIdentifiers(r dmarc.Record) string {
	var sb strings.Builder
	add := func(s string) {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(s)
	}

	section := func(title string) {
		add("  " + styleMuted.Render(title))
		add("")
	}
	field := func(label, val string) {
		add("  " + padRight(styleMuted.Render(label), 24) + styleWhite.Render(val))
	}

	section("MESSAGE IDENTIFIERS")
	field("Header From", r.Identifiers.HeaderFrom)
	// The XML's <spf><scope> states which address the checked domain came
	// from — "mfrom" (envelope-from / MAIL FROM) or "helo" (HELO/EHLO). We
	// label the domain accordingly rather than assuming "Envelope From",
	// since a "helo" scope names a different identifier entirely. When the
	// report omits <scope>, we name it only by what we know for certain:
	// the domain SPF was checked against.
	for _, s := range r.AuthResults.SPF {
		if s.Domain == "" {
			continue
		}
		label := "SPF Checked Domain"
		switch s.Scope {
		case "mfrom":
			label = "Envelope From"
		case "helo":
			label = "HELO/EHLO Domain"
		}
		field(label, s.Domain)
		break
	}
	add("")

	if len(r.AuthResults.DKIM) > 0 {
		section("DKIM SIGNATURE")
		for _, d := range r.AuthResults.DKIM {
			field("Signing domain (d=)", d.Domain)
			if d.Selector != "" {
				field("Selector (s=)", d.Selector)
			}
		}
	}

	return sb.String()
}

func (m model) renderDetailPolicy() string {
	fb := m.feedbacks[m.fileIdx]
	pol := fb.PolicyPublished

	orDash := func(s string) string {
		if s == "" {
			return "—"
		}
		return s
	}

	t := table.New().
		BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderHeader(false).
		StyleFunc(func(_, col int) lipgloss.Style {
			switch col {
			case 0:
				return styleCyan.PaddingRight(2)
			case 1:
				return styleWhite.PaddingRight(3)
			default:
				return styleMuted
			}
		}).
		Row("p", orDash(pol.P), "Policy applied to the domain").
		Row("sp", orDash(pol.SP), "Subdomain policy").
		Row("adkim", orDash(pol.ADKIM), "DKIM alignment mode (r = relaxed)").
		Row("aspf", orDash(pol.ASPF), "SPF alignment mode (r = relaxed)").
		Row("pct", fmt.Sprintf("%d%%", pol.PCT), "Percentage of messages filtered").
		Row("", "", "") // lipgloss/table v1.1.0: last row is always clipped by MaxHeight

	var sb strings.Builder
	fmt.Fprintf(&sb, "  %s\n\n", styleMuted.Render("PUBLISHED DMARC POLICY"))
	for line := range strings.SplitSeq(t.Render(), "\n") {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (m model) renderDetailRawXML(r dmarc.Record) string {
	type xmlWrapper struct {
		XMLName     xml.Name          `xml:"record"`
		Row         dmarc.Row         `xml:"row"`
		Identifiers dmarc.Identifiers `xml:"identifiers"`
		AuthResults dmarc.AuthResults `xml:"auth_results"`
	}

	wrapped := xmlWrapper{
		Row:         r.Row,
		Identifiers: r.Identifiers,
		AuthResults: r.AuthResults,
	}
	data, err := xml.MarshalIndent(wrapped, "", "    ")
	if err != nil {
		return styleRed.Render("error marshaling XML: " + err.Error())
	}

	var sb strings.Builder
	for i, line := range strings.Split(string(data), "\n") {
		if i > 0 {
			sb.WriteByte('\n')
		}
		fmt.Fprintf(&sb, "  %s  %s", styleMuted.Render(fmt.Sprintf("%3d", i+1)), colorizeXMLLine(line))
	}
	return sb.String()
}

// ── Detail tab helpers ────────────────────────────────────────────────────────

// mechDetail explains one mechanism's row. Its headline status is always
// row.policy_evaluated.{dkim,spf} — the DMARC-relevant verdict, the same
// field the hero's grid reads — never auth_results' raw check result.
// That distinction matters: a mechanism can authenticate successfully
// (auth_results says "pass") yet still fail DMARC because its domain
// doesn't align with the header-from. Showing the raw result as the
// headline here would print "✓ pass" under a row whose hero counterpart
// just said "✗ fail" for the same mechanism — a direct, confusing
// contradiction between two true-but-different numbers.
func mechDetail(rawResult, domain, selector, headerFrom string, verdictPass bool) string {
	if domain == "" {
		return styleMuted.Render("no signature presented")
	}
	parts := []string{"d=" + domain}
	if selector != "" {
		parts = append(parts, "s="+selector)
	}
	base := styleMuted.Render(strings.Join(parts, "  ·  "))
	switch {
	case verdictPass:
		return base + styleMuted.Render("  ·  ") + styleCyan.Render("aligned")
	case rawResult == "pass":
		return base + styleMuted.Render("  ·  authenticated, but ") +
			styleYellow.Render("doesn't align") +
			styleMuted.Render(" with "+headerFrom+" — doesn't count toward DMARC")
	default:
		return base + styleMuted.Render("  ·  ") + styleMuted.Render("not aligned")
	}
}

// renderMechRow writes a single mechanism line into sb. The detail text can
// be long, so it is word-wrapped at the available column width with
// continuation lines indented to align under the start of the detail —
// not back to column 0 like a naive table wrap.
func renderMechRow(sb *strings.Builder, label, verdict, detail string, W int) {
	const mechLabelW = 7 // "DKIM" (4) + PaddingRight(3), covers "SPF" (3) too
	paddedLabel := padRight(styleMuted.Render(label), mechLabelW)
	linePrefix := paddedLabel + verdict + "   "
	prefixW := lipgloss.Width(linePrefix)
	availW := max(W-4-2, 40) // box interior minus "  " indent
	detailW := max(availW-prefixW, 20)
	wrapped := lipgloss.NewStyle().Width(detailW).Render(detail)
	first := true
	for line := range strings.SplitSeq(wrapped, "\n") {
		line = strings.TrimRight(line, " ")
		if first {
			sb.WriteString("  " + linePrefix + line + "\n")
			first = false
		} else if line != "" {
			sb.WriteString(strings.Repeat(" ", 2+prefixW) + line + "\n")
		}
	}
}

// renderTableInto appends t's rendered rows to sb, each indented by two spaces.
// A dummy empty row is appended to t first to work around a lipgloss/table
// v1.1.0 off-by-one: computeHeight() returns N-1 rows, so MaxHeight clips
// the last real row — the dummy becomes the clipped one instead.
func renderTableInto(sb *strings.Builder, t *table.Table) {
	t.Row("", "")
	for line := range strings.SplitSeq(t.Render(), "\n") {
		sb.WriteString("  " + line + "\n")
	}
}
