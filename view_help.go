package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Help overlay ──────────────────────────────────────────────────────────────

// renderHelpView is a full-screen reference card, opened with "?" from
// anywhere in the app. It exists because the most cryptic things on screen
// aren't the data — they're the vocabulary: "aligned", "review", and the fact
// that DKIM/SPF appear as DMARC's aligned verdict rather than the raw
// mechanism check. Everything defined here is DMARC's own documented
// behavior (RFC 7489) — general reference knowledge, not a claim about any
// report's data — so it stays correct regardless of which file is open.
func (m model) renderHelpView() string {
	W := m.contentWidth()

	brand := lipgloss.NewStyle().Foreground(colBrand).Bold(true).Render("DMARC-TUI")
	left := " " + brand + styleMuted.Render(" | ") + styleWhite.Render("Help — reading the report")
	right := styleMuted.Render("press ") + styleWhiteBold.Render("esc") +
		styleMuted.Render(" or ") + styleWhiteBold.Render("?") +
		styleMuted.Render(" to close") + "  "
	header := fillLine(splitLine(left, right, W), W, colAppBg)

	h := m.help
	h.Width = W - 4
	footer := fillLine("  "+h.View(helpKeys), W, colAppBg)

	return header + "\n" + fillLine("", W, colAppBg) + "\n" + m.helpVP.View() + "\n" + footer
}

// renderHelpBody produces the scrollable content set into helpVP.
func (m model) renderHelpBody(W int) string {
	var sb strings.Builder
	fill := func(s string) { sb.WriteString(fillLine(s, W, colAppBg)); sb.WriteByte('\n') }
	blank := func() { fill("") }

	innerW := max(W-6, 40)

	// Section header: brand color, no underline — kept distinct from the blue
	// key/label column so headings and entries don't read as the same thing.
	section := func(title string) {
		blank()
		fill("  " + lipgloss.NewStyle().Foreground(colBrand).Bold(true).Render(title))
		blank()
	}

	// Key row: fixed-width key column so descriptions align.
	const keyColW = 18
	keyRow := func(keys, desc string) {
		fill("  " + padRight(styleAccentBold.Render(keys), keyColW) + styleWhite.Render(desc))
	}

	// Term: bold label + white body text. A blank line is added only when the
	// body wraps to more than one row — single-line entries don't need it.
	const termLabelW = 14
	term := func(label, body string) {
		bodyW := max(innerW-termLabelW, 20)
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(termLabelW).Render(styleAccentBold.Render(label)),
			styleWhite.Width(bodyW).Render(body),
		)
		lines := strings.Split(row, "\n")
		for _, line := range lines {
			fill("  " + line)
		}
		if len(lines) > 1 {
			blank()
		}
	}

	section("KEYS")
	keyRow("↑ ↓ / j k", "Move up / down in the list")
	keyRow("enter", "Open record detail")
	keyRow("esc", "Go back")
	keyRow("/", "Search by IP or domain")
	keyRow("f", "Cycle filter: All → Passing → Review → Failing")
	keyRow("s", "Sort by message volume or IP address")
	keyRow("[ ] · ← →", "Switch between report files")
	keyRow("tab · ← → · 1-5", "Switch detail section (Evaluation / Identifiers / Policy / Raw XML / IPInfo)")
	keyRow("pgup / pgdn", "Scroll detail content")
	keyRow("g / G", "Jump to top / bottom of list")
	keyRow("?", "Toggle this help")
	keyRow("q", "Quit")

	section("READING A RECORD'S VERDICT")
	term("DMARC", "Passes when DKIM or SPF authenticates the message and aligns with the From: address — the result that actually decides delivery.")
	term("DKIM / SPF", "Shown as the DMARC-aligned verdict, not the raw check. A mechanism can authenticate successfully yet still fail DMARC if its domain doesn't align with From:.")
	term("aligned", "The checked domain matches From: — exactly under strict mode, or via a shared organizational domain under relaxed mode (DMARC's default). E.g. mail.example.com aligns with news.example.com.")

	section("CLASSIFICATION")
	term("[ pass ]", "Both mechanisms aligned — a fully verified message.")
	term("[ review ]", "DMARC passed, but only one mechanism aligned — common for mail relayed through a third party (mailing list, ESP) under its own domain. Worth a glance, not necessarily wrong.")
	term("[ fail ]", "Neither mechanism aligned — DMARC failed for this message.")

	section("DISPOSITION")
	term("none", "Delivered and monitored only — no enforcement action taken.")
	term("quarantine", "Delivered to the spam / junk folder.")
	term("reject", "Rejected outright — never delivered to the recipient.")

	return strings.TrimRight(sb.String(), "\n")
}
