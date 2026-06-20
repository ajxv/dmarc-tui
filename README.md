# dmarc-tui

A terminal UI for browsing and making sense of DMARC aggregate reports
(the XML files mailbox providers send to your `rua=` address).

Point it at a report — or a folder of them — and get a navigable,
color-coded view of every record: who sent mail as your domain, whether
DKIM/SPF authenticated *and* aligned, what your policy did about it, and
why. A built-in glossary (press `?`) explains the vocabulary inline, so
you don't need DMARC memorized to read your own reports.

![dmarc-tui demo](docs/demo.gif)

## Install

**Quick install (Linux / macOS)** — downloads the right binary for your
platform from the latest release and puts it on your `PATH`
(`~/.local/bin` if writable, otherwise `/usr/local/bin`):

```sh
curl -fsSL https://raw.githubusercontent.com/ajxv/dmarc-tui/main/install.sh | sh
```

Run `install.sh --help` (or pipe `-s -- --help`) to see options like
`--version <tag>` and `--to <dir>`. You can also download the script
first and read it before running it, which is always a good idea for
anything piped into a shell:

```sh
curl -fsSL https://raw.githubusercontent.com/ajxv/dmarc-tui/main/install.sh -o install.sh
less install.sh        # read it
sh install.sh
```

**Download a prebuilt binary manually** from the [releases page](https://github.com/ajxv/dmarc-tui/releases) —
archives are published for Linux, macOS, and Windows (amd64 and arm64)
on every tagged release.

**Or build from source** (requires [Go](https://go.dev) 1.25+):

```sh
go install github.com/ajxv/dmarc-tui@latest
```

```sh
git clone https://github.com/ajxv/dmarc-tui.git
cd dmarc-tui
go build -o dmarc-tui .
```

## Usage

```sh
dmarc-tui <file.xml> [file2.xml …]   # one or more reports
dmarc-tui <directory>                # every *.xml in a directory
```

Reports are typically delivered as `.xml`, `.xml.gz`, or `.zip`
attachments by your mail provider — extract them to plain `.xml` first.

### Try it without your own data

A fully synthetic sample report ships in [sample-data/sample-report.xml](sample-data/sample-report.xml) —
every domain, IP, and org name in it is fictional (drawn from IANA's
reserved documentation ranges), so it's safe to poke at without
exposing real mail data:

```sh
dmarc-tui sample-data/sample-report.xml
```

## Keybindings

**Report list**

| Key | Action |
|-----|--------|
| `↑`/`k`, `↓`/`j` | Move the cursor |
| `g` / `G` | Jump to top / bottom |
| `enter` | Inspect the selected record |
| `/` | Search by source IP, authenticated domain, or From: address |
| `f` | Cycle the pass/fail filter |
| `s` | Cycle the sort order (volume / source IP) |
| `[`/`←`, `]`/`→` | Switch to the previous / next report file |
| `?` | Open the help & glossary screen |
| `q` | Quit |

**Record detail**

| Key | Action |
|-----|--------|
| `tab`, `←` / `→` | Cycle between tabs (forward / backward, wraps around) |
| `1`–`5` | Jump directly to a tab |
| `↑`/`k`, `↓`/`j` | Move to the previous / next record (without leaving detail view) |
| `pgup` / `pgdn` | Scroll the tab content |
| `?` | Open the help & glossary screen |
| `esc` / `q` | Back to the report list |

**Detail tabs**

| # | Tab | What it shows |
|---|-----|---------------|
| 1 | Evaluation | DMARC verdict, DKIM/SPF aligned results, disposition, classification |
| 2 | Identifiers | Header From, Envelope From / HELO domain, DKIM signing domain |
| 3 | Policy | Published DMARC policy (`p`, `sp`, `adkim`, `aspf`, `pct`) |
| 4 | Raw XML | The original record as indented XML |
| 5 | IPInfo | Live lookup from [ipinfo.io](https://ipinfo.io): hostname, location, ASN/org, timezone. Set `IPINFO_TOKEN` for a higher rate limit. |

The help screen (`?`) works as an overlay from anywhere — it explains
what "aligned" means, how to read the pass / review / fail
classification, and what each disposition (none / quarantine / reject)
means, then returns you to exactly where you were.

## Documentation

- [docs/glossary.md](docs/glossary.md) — DMARC concepts and how dmarc-tui presents them
- [docs/keybindings.md](docs/keybindings.md) — full keybinding reference

## Building

```sh
go build -o dmarc-tui .
```

Cross-platform release archives are attached automatically by
[.github/workflows/release.yml](.github/workflows/release.yml) whenever
a `vX.Y.Z` release is published on GitHub (Releases → Draft a new
release → tag `vX.Y.Z` → Publish).

## License

[MIT](LICENSE)
