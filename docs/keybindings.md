# Keybindings

dmarc-tui has three modes — the report list, a record's detail view, and
the help overlay — plus a small search input. Keys are scoped to
whichever of these is currently active.

## Report list

The default view: every record in the current report, sorted and
filterable.

| Key | Action |
|-----|--------|
| `↑` / `k`, `↓` / `j` | Move the cursor up / down |
| `g` | Jump to the first record |
| `G` | Jump to the last record |
| `enter` | Open the selected record's detail view |
| `/` | Start a search (matches source IP, authenticated domain, or From: address) |
| `f` | Cycle the filter: All → Passing → Review → Failing → All |
| `s` | Cycle the sort order: by volume → by source IP → by volume |
| `[` / `←` | Switch to the previous report file |
| `]` / `→` | Switch to the next report file |
| `?` | Open the help & glossary overlay |
| `q` / `ctrl+c` | Quit |

## Search

Entered with `/` from the report list.

| Key | Action |
|-----|--------|
| *(type)* | Filter records as you type |
| `enter` | Keep the current results and stop typing |
| `backspace` | Delete the last character |
| `esc` | Clear the search and return to the full list |

## Record detail

Opened with `enter` from the list — a focused view of one record across
five tabs.

| Key | Action |
|-----|--------|
| `tab` | Cycle to the next tab (wraps from IPInfo back to Evaluation) |
| `←` / `→` | Cycle backward / forward between tabs |
| `1`–`5` | Jump directly to a tab |
| `↑` / `k`, `↓` / `j` | Move to the previous / next record without leaving detail view |
| `pgup` / `pgdn` | Scroll the current tab's content |
| `?` | Open the help & glossary overlay |
| `esc` / `q` | Close detail view and return to the list |

### Tabs

| # | Name | Content |
|---|------|---------|
| 1 | Evaluation | DMARC aligned verdict, DKIM/SPF mechanism results, disposition, and record classification |
| 2 | Identifiers | Header From, Envelope From / HELO domain, DKIM signing domain and selector |
| 3 | Policy | The published DMARC policy: `p`, `sp`, `adkim`, `aspf`, `pct` |
| 4 | Raw XML | The original record rendered as indented, syntax-coloured XML |
| 5 | IPInfo | Live lookup from ipinfo.io: hostname, city, region, country, coordinates, ASN/org, postal code, timezone. Respects the `IPINFO_TOKEN` environment variable for authenticated requests (higher rate limit). Gracefully handles private/bogon IPs, network errors, and rate limits. |

## Help & glossary overlay

Opened with `?` from anywhere — a modal reference screen explaining the
keybindings and the vocabulary used throughout the app (DMARC, DKIM/SPF,
alignment, the pass / review / fail classification, and dispositions).
It closes back to exactly the screen it was opened from.

| Key | Action |
|-----|--------|
| `esc` / `?` / `q` | Close the overlay |
