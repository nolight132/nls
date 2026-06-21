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
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

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
		if block.Directory && (opts.Plain == PlainLong || opts.ShowBlocks) {
			total := blockTotal(block.Entries)
			if opts.Human {
				if _, err := fmt.Fprintf(w, "total %s\n", format.LsTotalSize(total*1024, true)); err != nil {
					return err
				}
			} else if _, err := fmt.Fprintf(w, "total %d\n", total); err != nil {
				return err
			}
		}
		if err := renderPlainEntries(w, block.Entries, opts, now); err != nil {
			return err
		}
	}
	return nil
}

func renderPlainEntries(w io.Writer, entries []listing.Entry, opts Options, now time.Time) error {
	switch opts.Plain {
	case PlainLong:
		widths := longWidths(entries, opts)
		for _, e := range entries {
			line := longLine(e, opts, now, widths)
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		return nil
	case PlainCommas:
		widths := plainWidths(entries, opts)
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, plainEntry(e, opts, widths))
		}
		_, err := fmt.Fprintln(w, strings.Join(names, ", "))
		return err
	default:
		widths := plainWidths(entries, opts)
		for _, e := range entries {
			name := plainEntry(e, opts, widths)
			if _, err := fmt.Fprintln(w, name); err != nil {
				return err
			}
		}
		return nil
	}
}

type longColumnWidths struct {
	inode  int
	blocks int
	links  int
	owner  int
	group  int
	size   int
}

type plainColumnWidths struct {
	inode  int
	blocks int
}

func plainWidths(entries []listing.Entry, opts Options) plainColumnWidths {
	widths := plainColumnWidths{}
	for _, e := range entries {
		if opts.ShowInode {
			widths.inode = max(widths.inode, len(fmt.Sprint(e.Inode)))
		}
		if opts.ShowBlocks {
			widths.blocks = max(widths.blocks, len(format.LsBlockSize(e.Blocks, opts.Human)))
		}
	}
	return widths
}

func plainEntry(e listing.Entry, opts Options, widths plainColumnWidths) string {
	parts := make([]string, 0, 3)
	if opts.ShowInode {
		parts = append(parts, fmt.Sprintf("%*d", widths.inode, e.Inode))
	}
	if opts.ShowBlocks {
		blocks := format.LsBlockSize(e.Blocks, opts.Human)
		parts = append(parts, fmt.Sprintf("%*s", widths.blocks, blocks))
	}
	parts = append(parts, listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName, false))
	return strings.Join(parts, " ")
}

func longWidths(entries []listing.Entry, opts Options) longColumnWidths {
	widths := longColumnWidths{links: 1, size: 1}
	for _, e := range entries {
		if opts.ShowInode {
			widths.inode = max(widths.inode, len(fmt.Sprint(e.Inode)))
		}
		if opts.ShowBlocks {
			widths.blocks = max(widths.blocks, len(format.LsBlockSize(e.Blocks, opts.Human)))
		}
		widths.links = max(widths.links, len(fmt.Sprint(e.Links)))
		widths.owner = max(widths.owner, len(e.Owner))
		widths.group = max(widths.group, len(e.Group))
		widths.size = max(widths.size, len(format.LsSize(e.Size, opts.Human, e.SizeApprox)))
	}
	return widths
}

func longLine(e listing.Entry, opts Options, now time.Time, widths longColumnWidths) string {
	var parts []string
	if opts.ShowInode {
		parts = append(parts, fmt.Sprintf("%*d", widths.inode, e.Inode))
	}
	if opts.ShowBlocks {
		blocks := format.LsBlockSize(e.Blocks, opts.Human)
		parts = append(parts, fmt.Sprintf("%*s", widths.blocks, blocks))
	}

	parts = append(parts, e.Permissions)
	parts = append(parts, fmt.Sprintf("%*d", widths.links, e.Links))
	parts = append(parts, fmt.Sprintf("%-*s", widths.owner, e.Owner))
	parts = append(parts, fmt.Sprintf("%-*s", widths.group, e.Group))

	size := format.LsSize(e.Size, opts.Human, e.SizeApprox)
	parts = append(parts, fmt.Sprintf("%*s", widths.size, size))

	when := format.LsTime(listing.EntryTime(e, opts.TimeField), now, opts.FullTime)
	parts = append(parts, when)

	name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName, true)
	parts = append(parts, name)

	return strings.Join(parts, " ")
}

func blockTotal(entries []listing.Entry) int64 {
	var total int64
	for _, e := range entries {
		total += e.Blocks
	}
	return total
}
