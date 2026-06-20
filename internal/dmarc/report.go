package dmarc

import (
	"net"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ── Classification ────────────────────────────────────────────────────────────

type Classification int

const (
	ClassPass Classification = iota
	ClassNeedsReview
	ClassFail
)

// Classify mirrors DMARC's own verdict rather than guessing at sender intent:
// row.policy_evaluated.{dkim,spf} are already the DMARC-aligned results, and
// the spec's pass condition is "dkim aligned-pass OR spf aligned-pass" — so
// both-pass and one-pass are equally a DMARC PASS. We still surface the
// one-mechanism case as "needs review" (worth a glance at *why* only one
// aligned) without implying it's a lesser or untrusted outcome.
func Classify(r Record) Classification {
	dk := r.Row.PolicyEvaluated.DKIM
	sp := r.Row.PolicyEvaluated.SPF
	if dk == "pass" && sp == "pass" {
		return ClassPass
	}
	if dk == "pass" || sp == "pass" {
		return ClassNeedsReview
	}
	return ClassFail
}

// AuthDomain returns the domain the report itself names as responsible for
// the message: the DKIM signing domain (the d= tag) when a signature is
// present, else the domain SPF was checked against. Bare IP addresses in the
// SPF domain field are skipped — they aren't a domain in the "authenticated
// by [domain]" sense and would read as a glitch in the UI.
func AuthDomain(r Record) string {
	if len(r.AuthResults.DKIM) > 0 && r.AuthResults.DKIM[0].Domain != "" {
		return r.AuthResults.DKIM[0].Domain
	}
	if len(r.AuthResults.SPF) > 0 && r.AuthResults.SPF[0].Domain != "" {
		if d := r.AuthResults.SPF[0].Domain; net.ParseIP(d) == nil {
			return d
		}
	}
	return ""
}

// IsAligned reports DMARC identifier alignment per RFC 7489 §3.1: "strict"
// (mode "s") requires an exact match; "relaxed" (mode "r", the default)
// requires only a shared Organizational Domain per the Public Suffix List.
func IsAligned(authDom, headerFrom, mode string) bool {
	authDom = strings.ToLower(strings.TrimSpace(authDom))
	headerFrom = strings.ToLower(strings.TrimSpace(headerFrom))
	if authDom == "" || headerFrom == "" {
		return false
	}
	if authDom == headerFrom {
		return true
	}
	if mode == "s" {
		return false
	}
	authOrg, errA := publicsuffix.EffectiveTLDPlusOne(authDom)
	fromOrg, errF := publicsuffix.EffectiveTLDPlusOne(headerFrom)
	if errA != nil || errF != nil {
		return false
	}
	return authOrg == fromOrg
}

// RecordMatches reports whether r's source IP, authenticated domain, or
// header-from address contains the (already lower-cased) search query.
func RecordMatches(r Record, q string) bool {
	return strings.Contains(strings.ToLower(r.Row.SourceIP), q) ||
		strings.Contains(strings.ToLower(AuthDomain(r)), q) ||
		strings.Contains(strings.ToLower(r.Identifiers.HeaderFrom), q)
}
