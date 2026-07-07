package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/nolight132/nls/internal/format"
	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
	"golang.org/x/term"
)

// PlainMode selects piped/plain output layout.
type PlainMode int

const (
	PlainOne PlainMode = iota
	PlainLong
	PlainCommas
)

// RenderOptions control rendered output.
type RenderOptions struct {
	Human      bool
	Long       bool
	JSON       bool
	Color      bool
	IconSet    icons.Set
	IsTTY      bool
	Now        time.Time
	Plain      PlainMode
	Classify   bool
	DirSlash   bool
	QuoteName  bool
	ShowInode  bool
	ShowBlocks bool
	UseTable   bool
	// Width caps table rendering at the terminal width. 0 means no cap.
	Width int
	// Columns controls which columns appear in table mode and their order.
	// Empty falls back to the built-in default set.
	Columns []string
}

// JSONRow is a single entry in JSON output.
type JSONRow struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	SizeHuman   string `json:"size_human,omitempty"`
	Modified    string `json:"modified,omitempty"`
	Permissions string `json:"permissions"`
	LinkTarget  string `json:"link_target,omitempty"`
}

// Render writes listing output according to options.
func Render(w io.Writer, blocks []listing.Block, opts RenderOptions) error {
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}

	if opts.JSON {
		return renderJSON(w, blocks, opts)
	}
	if opts.UseTable {
		for i, block := range blocks {
			if i > 0 {
				fmt.Fprintln(w)
			}
			if block.Header != "" {
				fmt.Fprintf(w, "%s:\n", block.Header)
			}
			if err := renderTable(w, block.Entries, opts); err != nil {
				return err
			}
		}
		return nil
	}
	return renderPlain(w, blocks, opts)
}

func renderJSON(w io.Writer, blocks []listing.Block, opts RenderOptions) error {
	var rows []JSONRow
	for _, b := range blocks {
		for _, e := range b.Entries {
			row := JSONRow{
				Name:        e.Name,
				Path:        entryPath(b.Dir, e.Name),
				Type:        typeLabel(e),
				Size:        e.Size,
				Permissions: e.Permissions,
				LinkTarget:  e.LinkTarget,
			}
			if !e.Modified.IsZero() {
				row.Modified = e.Modified.Format(time.RFC3339)
			}
			if opts.Human {
				row.SizeHuman = format.Size(e.Size, true, e.SizeApprox)
			}
			rows = append(rows, row)
		}
	}
	if rows == nil {
		rows = []JSONRow{}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

func entryPath(dir, name string) string {
	if dir == "" || name == "." || name == ".." {
		return name
	}
	return filepath.Join(dir, name)
}

func typeLabel(e listing.Entry) string {
	switch e.Kind {
	case listing.KindDirectory:
		return "dir"
	case listing.KindSymlink:
		return "link"
	case listing.KindExecutable:
		return "exec"
	default:
		return "file"
	}
}

// StdoutIsTTY reports whether stdout is a terminal.
func StdoutIsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// WriteError writes an ls-style error to stderr.
func WriteError(err error, suggest bool) {
	msg := formatError(err, suggest)
	fmt.Fprintln(os.Stderr, strings.TrimSpace(msg))
}

func formatError(err error, suggest bool) string {
	if pathErr, ok := errors.AsType[*os.PathError](err); ok {
		path := displayErrorPath(err, pathErr)
		msg := fmt.Sprintf("nls: %s: %s", path, sentenceCase(pathErr.Err.Error()))
		if suggest && errors.Is(pathErr.Err, os.ErrNotExist) {
			if candidate := didYouMean(path); candidate != "" {
				msg += fmt.Sprintf("\nDid you mean '%s'?", candidate)
			}
		}
		return msg
	}
	return "nls: " + strings.TrimSpace(err.Error())
}

func displayErrorPath(err error, pathErr *os.PathError) string {
	path := pathErr.Path
	if prefix, _, ok := strings.Cut(err.Error(), ": "); ok && prefix != "" && !strings.HasPrefix(prefix, pathErr.Op+" ") {
		path = prefix
	}
	return path
}

func didYouMean(path string) string {
	dir, base := filepath.Split(path)
	searchDir := dir
	if searchDir == "" {
		searchDir = "."
	}
	entries, err := os.ReadDir(searchDir)
	if err != nil || base == "" {
		return ""
	}

	bestName := ""
	bestDistance := 0
	for _, entry := range entries {
		name := entry.Name()
		distance := levenshtein(strings.ToLower(base), strings.ToLower(name))
		if bestName == "" || distance < bestDistance {
			bestName = name
			bestDistance = distance
		}
	}
	if bestName == "" || bestDistance > suggestionDistance(base) {
		return ""
	}
	return dir + bestName
}

func suggestionDistance(s string) int {
	limit := len([]rune(s)) / 3
	if limit < 2 {
		return 2
	}
	return limit
}

func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i, ra := range ar {
		curr := make([]int, len(br)+1)
		curr[0] = i + 1
		for j, rb := range br {
			cost := 0
			if ra != rb {
				cost = 1
			}
			curr[j+1] = minInt(curr[j]+1, prev[j+1]+1, prev[j]+cost)
		}
		prev = curr
	}
	return prev[len(br)]
}

func minInt(values ...int) int {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}

// sanitizeName hides control characters on terminals, where a crafted
// filename could break table layout or inject escape sequences.
func sanitizeName(s string) string {
	if !strings.ContainsFunc(s, unicode.IsControl) {
		return s
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '?'
		}
		return r
	}, s)
}

func sentenceCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func StderrIsTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}
