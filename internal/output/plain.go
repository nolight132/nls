package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nolight132/nls/internal/format"
	"github.com/nolight132/nls/internal/listing"
)

func renderPlain(w io.Writer, blocks []listing.Block, opts Options) error {
	for bi, block := range blocks {
		if bi > 0 && opts.Plain != PlainCommas {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if block.Header != "" {
			if _, err := fmt.Fprintf(w, "%s:\n", block.Header); err != nil {
				return err
			}
		}
		if err := renderPlainEntries(w, block.Entries, opts); err != nil {
			return err
		}
	}
	return nil
}

func renderPlainEntries(w io.Writer, entries []listing.Entry, opts Options) error {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	switch opts.Plain {
	case PlainLong:
		for _, e := range entries {
			line := longLine(e, opts, now)
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		return nil
	case PlainCommas:
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName))
		}
		_, err := fmt.Fprintln(w, strings.Join(names, ", "))
		return err
	default:
		for _, e := range entries {
			name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName)
			if _, err := fmt.Fprintln(w, name); err != nil {
				return err
			}
		}
		return nil
	}
}

func longLine(e listing.Entry, opts Options, now time.Time) string {
	var parts []string
	if opts.ShowInode && e.Inode > 0 {
		parts = append(parts, fmt.Sprintf("%7d", e.Inode))
	}
	if opts.ShowBlocks {
		parts = append(parts, fmt.Sprintf("%4d", e.Blocks))
	}

	parts = append(parts, e.Permissions)
	parts = append(parts, "1", "-", "-")

	size := format.Size(e.Size, opts.Human, e.SizeApprox)
	parts = append(parts, fmt.Sprintf("%8s", size))

	when := format.LsTime(listing.EntryTime(e, opts.TimeField), now, opts.FullTime)
	parts = append(parts, when)

	name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName)
	parts = append(parts, name)

	return strings.Join(parts, " ")
}
