# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [v1.0.0]

First tagged release.

### Added

- Report list view: searchable, filterable (All / Passing / Review /
  Failing), sortable (volume / source IP) table of every record in a
  DMARC aggregate report, with at-a-glance pass-rate and DMARC/DKIM/SPF
  stat panels.
- Record detail view with five tabs — Evaluation, Identifiers, Policy,
  Raw XML, and IPInfo — reachable via `enter`, `tab`, `←`/`→`, or number
  keys `1`–`5`.
- Evaluation tab that walks through *why* a record passed or failed:
  the DMARC-aligned verdict for DKIM and SPF, the raw mechanism result,
  the domain each was checked against, and whether it aligned with the
  message's `From:` domain (strict or relaxed, per RFC 7489 §3.1).
- Pass / review / fail classification that mirrors DMARC's own pass
  condition (DKIM-aligned-pass OR SPF-aligned-pass), rather than
  requiring both to succeed.
- Detail view hero section: icon badge (✓/✗), labelled Source IP /
  Sending domain / Email From fields with message count and date, and
  a right-aligned DMARC/DKIM/SPF status column that stays at a fixed
  position across terminal widths.
- IPInfo tab: live lookup from ipinfo.io for any source IP — hostname,
  city, region, country, coordinates, ASN/org, postal code, and timezone.
  Results are cached per-IP for the session. Handles private/bogon
  addresses, network errors, timeouts, and rate limits with specific
  messages. Respects `IPINFO_TOKEN` for the authenticated endpoint.
- Left/right arrow keys in detail view navigate between tabs (with
  wrap-around); `tab` still cycles forward.
- In-app help & glossary overlay (`?`) explaining the keybindings and
  the vocabulary used throughout — DMARC, DKIM/SPF, alignment, the
  classification badges, and dispositions — without leaving the screen
  you're on.
- Multi-report navigation (`[` / `]`) for browsing several files or an
  entire directory of reports in one session.
- Responsive layout that adapts to the terminal width: column widths
  scale between defined minimums and maximums; a minimum-size guard
  (100 × 20) shows a clear "resize terminal" message instead of a
  broken layout.
- Tokyo Night–themed UI built on Bubble Tea / Lip Gloss.
- `--version` / `-v` and `--help` / `-h` flags.
- A fully synthetic sample report (`sample-data/sample-report.xml`,
  built entirely from IANA-reserved documentation domains and addresses)
  for trying out the tool without real mail data.
- Cross-platform release builds (Linux / macOS / Windows, amd64 / arm64)
  via GoReleaser, published automatically on tagged releases.

[v1.0.0]: https://github.com/ajxv/dmarc-tui/releases/tag/v1.0.0
