package ipinfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Data mirrors the fields ipinfo.io returns in its JSON response.
// Bogon is set for private / reserved addresses that have no public record.
type Data struct {
	IP       string `json:"ip"`
	Bogon    bool   `json:"bogon"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

// CacheEntry stores the result of one ipinfo.io lookup.
// Done=false means the fetch is still in flight.
type CacheEntry struct {
	Done bool
	Info *Data  // non-nil on success
	Err  string // non-empty on error
}

// Msg is the tea.Msg delivered when a fetch completes (success or error).
type Msg struct {
	IP   string
	Info *Data
	Err  string
}

// Fetch returns a tea.Cmd that queries ipinfo.io for ip and delivers a Msg.
// Set IPINFO_TOKEN in the environment for the authenticated endpoint
// (50 k req/month free; a token raises the limit).
func Fetch(ip string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}

		endpoint := "https://ipinfo.io/" + ip + "/json"
		if tok := os.Getenv("IPINFO_TOKEN"); tok != "" {
			endpoint += "?token=" + tok
		}

		resp, err := client.Get(endpoint)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) && urlErr.Timeout() {
				return Msg{IP: ip, Err: "request timed out (5 s)"}
			}
			return Msg{IP: ip, Err: "network unavailable — check your connection"}
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return Msg{IP: ip, Err: "rate limit exceeded (free tier: 50 k req/month — set IPINFO_TOKEN for a higher limit)"}
		case http.StatusUnauthorized, http.StatusForbidden:
			return Msg{IP: ip, Err: "invalid token — check IPINFO_TOKEN"}
		}
		if resp.StatusCode != http.StatusOK {
			return Msg{IP: ip, Err: fmt.Sprintf("HTTP %d from ipinfo.io", resp.StatusCode)}
		}

		var data Data
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return Msg{IP: ip, Err: "unexpected response format"}
		}
		return Msg{IP: ip, Info: &data}
	}
}
