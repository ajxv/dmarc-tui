package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"dmarc-tui/internal/dmarc"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z" by the
// release workflow; local `go build` runs report "dev".
var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: dmarc-tui <file.xml> [file2.xml …]  OR  dmarc-tui <directory>")
		os.Exit(1)
	}

	switch args[0] {
	case "-v", "--version":
		fmt.Println("dmarc-tui " + version)
		return
	case "-h", "--help":
		fmt.Println("Usage: dmarc-tui <file.xml> [file2.xml …]  OR  dmarc-tui <directory>")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -v, --version   print the version and exit")
		fmt.Println("  -h, --help      show this help and exit")
		return
	}

	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		if info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(arg, "*.xml"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error scanning %s: %v\n", arg, err)
				continue
			}
			sort.Strings(matches)
			files = append(files, matches...)
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No XML files found.")
		os.Exit(1)
	}

	var feedbacks []*dmarc.Feedback
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", f, err)
			continue
		}
		var fb dmarc.Feedback
		if err := xml.Unmarshal(data, &fb); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing %s: %v\n", f, err)
			continue
		}
		feedbacks = append(feedbacks, &fb)
	}

	if len(feedbacks) == 0 {
		fmt.Fprintln(os.Stderr, "No valid DMARC reports found.")
		os.Exit(1)
	}

	p := tea.NewProgram(newModel(feedbacks), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
