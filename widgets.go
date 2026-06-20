package main

import (
	"fmt"
	"strings"

	"dmarc-tui/internal/dmarc"
	"github.com/charmbracelet/lipgloss"
)

// ── Authenticated domain color ────────────────────────────────────────────────

// domainStyled renders the authenticated domain in cyan — or, for the
// selected row (bright), a bolder cyan plus a brighter dash placeholder so
// the whole row reads as "lit up" without relying on a background fill.
func domainStyled(domain string, bright bool) string {
	if domain == "" {
		if bright {
			return styleWhite.Render("—")
		}
		return styleMuted.Render("—")
	}
	if bright {
		return styleCyanBold.Render(domain)
	}
	return styleCyan.Render(domain)
}

// ── Outline badges ────────────────────────────────────────────────────────────

// outlineBadge renders "[ text ]" in a single accent color — an outline-style
// tag with no background fill, padded to a fixed inner width when innerW > 0.
func outlineBadge(text string, fg lipgloss.Color, innerW int) string {
	if pad := innerW - lipgloss.Width(text); pad > 0 {
		lpad := pad / 2
		text = strings.Repeat(" ", lpad) + text + strings.Repeat(" ", pad-lpad)
	}
	br := lipgloss.NewStyle().Foreground(fg)
	txt := lipgloss.NewStyle().Foreground(fg).Bold(true)
	return br.Render("[ ") + txt.Render(text) + br.Render(" ]")
}

// ── Classification badge ──────────────────────────────────────────────────────

func classBadge(c dmarc.Classification) string {
	var text string
	var clr lipgloss.Color
	switch c {
	case dmarc.ClassPass:
		text, clr = "pass", colGreen
	case dmarc.ClassNeedsReview:
		text, clr = "review", colYellow
	default:
		text, clr = "fail", colRed
	}
	return outlineBadge(text, clr, classBadgeW-4)
}

// ── Auth inline ───────────────────────────────────────────────────────────────

// authInline renders the DMARC/DKIM/SPF check row. For the selected row
// (bright) the labels switch from muted to white so the whole line "lights
// up" in place of a background fill.
func authInline(dmarcOk bool, dk, sp string, bright bool) string {
	lbl := styleMuted
	if bright {
		lbl = styleWhite
	}
	chk := func(label string, ok bool) string {
		if ok {
			return styleGreen.Render("✓") + " " + lbl.Render(label)
		}
		return styleRed.Render("✗") + " " + lbl.Render(label)
	}
	return chk("DMARC", dmarcOk) + "  " + chk("DKIM", dk == "pass") + "  " + chk("SPF", sp == "pass")
}

// ── Status indicators ─────────────────────────────────────────────────────────

// statusPassFail renders a pass/fail indicator for the detail hero.
func statusPassFail(ok bool) string {
	if ok {
		return styleGreenBold.Render("✓ pass")
	}
	return styleRedBold.Render("✗ fail")
}

// resultStr renders a DMARC/DKIM/SPF result word as a colored glyph + text.
// The brighter pass/fail colors (colPassBright, colFailBright) are distinct
// from the softer badge palette — they carry more visual weight on a plain
// background and make the verdict read clearly at a glance.
func resultStr(result string) string {
	var glyph string
	var fg lipgloss.Color
	switch strings.ToLower(result) {
	case "pass":
		glyph, fg = "✓", colPassBright
	case "fail":
		glyph, fg = "✗", colFailBright
	default:
		glyph, fg = "○", colMuted
	}
	return lipgloss.NewStyle().Foreground(fg).Render(glyph + " " + strings.ToLower(result))
}

// dispositionGloss pairs the policy's raw enum value with a plain-language
// note, since "none"/"quarantine"/"reject" name an action the report assumes
// the reader already knows how to interpret.
func dispositionGloss(d string) string {
	var note string
	switch strings.ToLower(d) {
	case "none":
		note = "no action taken — monitored only"
	case "quarantine":
		note = "delivered to spam / junk"
	case "reject":
		note = "rejected outright"
	}
	val := styleWhite.Render(d)
	if note != "" {
		val += styleMuted.Render("   " + note)
	}
	return val
}

// ── Volume bar ────────────────────────────────────────────────────────────────

func volumeBar(val, total, width int) string {
	if total == 0 || width <= 0 {
		return styleDimMuted.Render(strings.Repeat("░", width))
	}
	filled := min(int(float64(val)/float64(total)*float64(width)+0.5), width)
	bar := lipgloss.NewStyle().Foreground(colAccent).Render(strings.Repeat("█", filled))
	rest := styleDimMuted.Render(strings.Repeat("░", width-filled))
	return bar + rest
}

// ── Pass-rate bar (stat panel) ────────────────────────────────────────────────

func lineBar(val, total, width int) string {
	if total == 0 || width <= 0 {
		return styleMuted.Render(strings.Repeat("─", width))
	}
	filled := min(int(float64(val)/float64(total)*float64(width)), width)
	p := float64(val) / float64(total) * 100
	var fill lipgloss.Style
	switch {
	case p >= 95:
		fill = styleGreen
	case p >= 75:
		fill = styleYellow
	default:
		fill = styleRed
	}
	return fill.Render(strings.Repeat("━", filled)) + styleDimMuted.Render(strings.Repeat("─", width-filled))
}

func passRateBadge(p float64) string {
	s := fmt.Sprintf("%.1f%%", p)
	switch {
	case p >= 95:
		return styleGreenBold.Render(s)
	case p >= 75:
		return styleYellowBold.Render(s)
	default:
		return styleRedBold.Render(s)
	}
}

// ── XML colorizer ─────────────────────────────────────────────────────────────

func colorizeXMLLine(line string) string {
	var out strings.Builder
	for len(line) > 0 {
		switch {
		case strings.HasPrefix(line, "</"):
			end := strings.Index(line, ">")
			if end < 0 {
				out.WriteString(styleMuted.Render(line))
				return out.String()
			}
			out.WriteString(styleMuted.Render("</"))
			out.WriteString(styleAccent.Render(line[2:end]))
			out.WriteString(styleMuted.Render(">"))
			line = line[end+1:]
		case strings.HasPrefix(line, "<"):
			end := strings.Index(line, ">")
			if end < 0 {
				out.WriteString(styleMuted.Render(line))
				return out.String()
			}
			out.WriteString(styleMuted.Render("<"))
			out.WriteString(styleAccent.Render(line[1:end]))
			out.WriteString(styleMuted.Render(">"))
			line = line[end+1:]
		default:
			next := strings.Index(line, "<")
			if next < 0 {
				out.WriteString(styleWhite.Render(line))
				return out.String()
			}
			out.WriteString(styleWhite.Render(line[:next]))
			line = line[next:]
		}
	}
	return out.String()
}
