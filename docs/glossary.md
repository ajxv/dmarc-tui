# Reading a DMARC report

DMARC aggregate reports are blunt instruments: a wall of IPs, domains,
and three-letter verdicts with no narrative. The data is precise, but
the *vocabulary* is the part that trips people up — "pass" can mean two
different things depending on which field you're looking at, and a
record can look self-contradictory until you know what to check next.

This page explains the concepts dmarc-tui's screens are built around. A
condensed version of the same material is available in the app itself —
press `?` from anywhere.

## The two checks DMARC builds on

Every record carries the results of two independent authentication
mechanisms:

- **SPF** (Sender Policy Framework) — checks whether the sending mail
  server's IP is authorized to send for the domain it claims in the
  SMTP "envelope" (`MAIL FROM` / HELO).
- **DKIM** (DomainKeys Identified Mail) — checks a cryptographic
  signature attached to the message, which names the domain that signed
  it (the `d=` tag).

Both can succeed *as authentication checks* — the IP really is
authorized, the signature really does verify — and DMARC can still
fail. That's because DMARC adds a second condition on top: **alignment**.

## Alignment

A passing SPF or DKIM check only counts toward DMARC if the domain it
authenticated **aligns** with the domain in the message's visible
`From:` header — the address a recipient actually sees. Two modes
control how strict that match has to be (set per-domain in the
sender's published DMARC policy, `adkim`/`aspf`):

- **strict** (`s`) — the domains must match exactly.
- **relaxed** (`r`, the default when unset) — it's enough for both
  domains to share the same *Organizational Domain*, the registrable
  root found via the Public Suffix List. `mail.example.com` and
  `news.example.com` both reduce to `example.com`, so they align under
  relaxed mode even though neither matches the other (or `example.com`
  itself) exactly.

This is why a record can show, say, "SPF: pass" in one place and "SPF:
fail" in another and *both be correct* — one is the raw mechanism
result, the other is the DMARC-aligned verdict. dmarc-tui always leads
with the aligned verdict (what actually decided the message's fate) and
shows the raw result alongside it for context.

## DMARC's pass condition

A message passes DMARC if **either** DKIM or SPF authenticates *and*
aligns. It doesn't need both. That single rule explains most of the
"why does this look weird" moments in a report:

- **Both aligned** — a fully verified message, sent directly by (or on
  behalf of, under its own name) the domain in `From:`.
- **Only one aligned** — DMARC still passes, but it's worth a glance at
  *why* the other didn't. The most common reason is mail relayed
  through a third party — a mailing list, forwarder, or marketing/ESP
  platform — that signs or sends under its own domain rather than
  yours. SPF often breaks first in this scenario (the relay is the
  envelope sender), while your original DKIM signature can survive the
  trip and keep the message aligned and passing.
- **Neither aligned** — DMARC fails for this message, regardless of
  what the raw SPF/DKIM results say. This is the case your policy
  (`p=`) acts on.

dmarc-tui reflects this directly in its classification badges:

| Badge | Meaning |
|-------|---------|
| `pass` | Both DKIM and SPF authenticated and aligned. |
| `review` | DMARC passed (one mechanism aligned), but the other didn't — usually a relay/ESP pattern, not necessarily a problem. |
| `fail` | Neither mechanism aligned; DMARC failed outright. |

## Disposition — what the receiver actually did

`policy_evaluated.disposition` records the action the *receiving* mail
system took as a result of the DMARC evaluation, per the policy your
domain published:

| Disposition | Meaning |
|-------------|---------|
| `none` | Delivered normally — the policy only requested monitoring (`p=none`), or the message passed. |
| `quarantine` | Delivered to spam/junk. |
| `reject` | Rejected outright; never delivered. |

## A field worth knowing about: identifiers vs. auth results

A record's `identifiers.header_from` is the domain that ends up in the
`From:` address recipients see — the thing alignment is measured
against. `auth_results` carries the raw, independent SPF and DKIM
findings (which domains were checked, what they returned). DMARC's
`policy_evaluated` then combines all three: did either mechanism
authenticate *for a domain that aligns with header_from*? dmarc-tui's
Evaluation tab walks through exactly this chain for any record you
inspect, so you can see which piece changed the outcome.

## Further reading

This page sticks to what you need to read dmarc-tui's own screens. For
the full specification, see [RFC 7489](https://www.rfc-editor.org/rfc/rfc7489).
