package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/termcolor"
)

func renderPlain(w io.Writer, blocks []listing.Block, opts RenderOptions) error {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	styles := termcolor.New(opts.Color)

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
		blockOpts := opts
		blockOpts.Columns = columnsForBlock(block, opts.Columns)
		if err := renderPlainBlock(w, block.Entries, blockOpts, now, styles); err != nil {
			return err
		}
	}
	return nil
}

func renderPlainBlock(w io.Writer, entries []listing.Entry, opts RenderOptions, now time.Time, styles *termcolor.Style) error {
	switch opts.Plain {
	case PlainCommas:
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, plainName(e, opts))
		}
		_, err := fmt.Fprintln(w, strings.Join(names, ", "))
		return err
	case PlainLong:
		return renderPlainColumns(w, entries, opts, now, styles)
	default:
		for _, e := range entries {
			if _, err := fmt.Fprintln(w, plainName(e, opts)); err != nil {
				return err
			}
		}
		return nil
	}
}

func plainName(e listing.Entry, opts RenderOptions) string {
	name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName, false)
	if opts.IsTTY {
		name = sanitizeName(name)
	}
	// ls prefixes "inode blocks name" in every output format.
	if opts.ShowBlocks {
		name = fmt.Sprintf("%d %s", e.Blocks, name)
	}
	if opts.ShowInode {
		name = fmt.Sprintf("%d %s", e.Inode, name)
	}
	return name
}

// renderPlainColumns renders the same columns the table would show, aligned
// as plain text without borders or headers.
func renderPlainColumns(w io.Writer, entries []listing.Entry, opts RenderOptions, now time.Time, styles *termcolor.Style) error {
	cols := buildTableColumns(opts, styles)
	if len(cols) == 0 {
		return nil
	}

	ctx := renderCtx{opts: opts, styles: styles, now: now, human: opts.Human || opts.IsTTY}

	rows := make([][]string, 0, len(entries))
	for i, e := range entries {
		row := make([]string, 0, len(cols))
		for _, col := range cols {
			row = append(row, col.render(e, i, ctx))
		}
		rows = append(rows, row)
	}
	rowWidths := measureRows(rows)

	widths := make([]int, len(cols))
	for ri := range rows {
		for i, cw := range rowWidths[ri] {
			if cw > widths[i] {
				widths[i] = cw
			}
		}
	}

	for ri, row := range rows {
		var b strings.Builder
		for i, cell := range row {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(alignCell(cell, rowWidths[ri][i], widths[i], cols[i].align))
		}
		b.WriteByte('\n')
		if _, err := w.Write([]byte(b.String())); err != nil {
			return err
		}
	}
	return nil
}
