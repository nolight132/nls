package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/nolight132/nls/internal/format"
	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/termcolor"
)

type cellAlign int

const (
	alignLeft cellAlign = iota
	alignRight
	alignCenter
)

type tableColumn struct {
	header string
	align  cellAlign
}

func renderTable(w io.Writer, entries []listing.Entry, opts Options) error {
	styles := termcolor.New(opts.Color)
	cols := []tableColumn{
		{header: styles.Header("#"), align: alignRight},
		{header: styles.Header("name"), align: alignLeft},
		{header: styles.Header("type"), align: alignLeft},
		{header: styles.Header("size"), align: alignRight},
		{header: styles.Header("modified"), align: alignLeft},
	}
	if opts.ShowInode {
		cols = append(cols, tableColumn{header: styles.Header("inode"), align: alignRight})
	}
	if opts.ShowBlocks {
		cols = append(cols, tableColumn{header: styles.Header("blocks"), align: alignRight})
	}
	if opts.Long {
		cols = append(cols, tableColumn{header: styles.Header("permissions"), align: alignLeft})
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	human := opts.Human || opts.IsTTY
	rows := make([][]string, 0, len(entries))
	for i, e := range entries {
		name := tableDisplayName(e, opts)
		name = styles.Name(name, e.Kind)

		modified := tableTime(e, opts, now)
		modified = styles.Modified(modified)

		row := []string{
			styles.Index(strconv.Itoa(i)),
			name,
			typeLabel(e),
			styles.Size(format.Size(e.Size, human, e.SizeApprox)),
			modified,
		}
		if opts.ShowInode {
			row = append(row, strconv.FormatUint(e.Inode, 10))
		}
		if opts.ShowBlocks {
			row = append(row, strconv.FormatInt(e.Blocks, 10))
		}
		if opts.Long {
			row = append(row, e.Permissions)
		}
		rows = append(rows, row)
	}

	table := buildBorderedTable(cols, rows)
	_, err := fmt.Fprint(w, table)
	return err
}

func tableDisplayName(e listing.Entry, opts Options) string {
	name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName, true)
	return icons.For(e.Kind, opts.IconSet) + name
}

func tableTime(e listing.Entry, opts Options, now time.Time) string {
	t := listing.EntryTime(e, opts.TimeField)
	if opts.FullTime {
		return format.LsTime(t, now, true)
	}
	return format.Modified(t, now)
}

func buildBorderedTable(cols []tableColumn, rows [][]string) string {
	widths := columnWidths(cols, rows)

	var b strings.Builder
	writeBorderTop(&b, widths)
	writeHeaderRow(&b, widths, cols)
	writeBorderMid(&b, widths)
	aligns := rowAligns(cols)
	for _, row := range rows {
		writeDataRow(&b, widths, row, aligns)
	}
	writeBorderBottom(&b, widths)
	return b.String()
}

func columnWidths(cols []tableColumn, rows [][]string) []int {
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = visibleWidth(col.header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if w := visibleWidth(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func writeBorderTop(b *strings.Builder, widths []int) {
	b.WriteString("╭")
	for i, w := range widths {
		if i > 0 {
			b.WriteString("┬")
		}
		b.WriteString(strings.Repeat("─", w+2))
	}
	b.WriteString("╮\n")
}

func writeBorderMid(b *strings.Builder, widths []int) {
	b.WriteString("├")
	for i, w := range widths {
		if i > 0 {
			b.WriteString("┼")
		}
		b.WriteString(strings.Repeat("─", w+2))
	}
	b.WriteString("┤\n")
}

func writeBorderBottom(b *strings.Builder, widths []int) {
	b.WriteString("╰")
	for i, w := range widths {
		if i > 0 {
			b.WriteString("┴")
		}
		b.WriteString(strings.Repeat("─", w+2))
	}
	b.WriteString("╯\n")
}

func writeHeaderRow(b *strings.Builder, widths []int, cols []tableColumn) {
	b.WriteRune('│')
	for i, col := range cols {
		if i > 0 {
			b.WriteRune('│')
		}
		align := col.align
		switch stripANSI(col.header) {
		case "#", "name", "size", "modified":
			align = alignCenter
		}
		b.WriteString(" ")
		b.WriteString(alignCell(col.header, widths[i], align))
		b.WriteString(" ")
	}
	b.WriteString("│\n")
}

func writeDataRow(b *strings.Builder, widths []int, row []string, aligns []cellAlign) {
	b.WriteRune('│')
	for i, cell := range row {
		if i > 0 {
			b.WriteRune('│')
		}
		b.WriteString(" ")
		b.WriteString(alignCell(cell, widths[i], aligns[i]))
		b.WriteString(" ")
	}
	b.WriteString("│\n")
}

func rowAligns(cols []tableColumn) []cellAlign {
	aligns := make([]cellAlign, len(cols))
	for i, col := range cols {
		aligns[i] = col.align
	}
	return aligns
}

func alignCell(cell string, width int, align cellAlign) string {
	pad := width - visibleWidth(cell)
	if pad <= 0 {
		return cell
	}

	switch align {
	case alignRight:
		return strings.Repeat(" ", pad) + cell
	case alignCenter:
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + cell + strings.Repeat(" ", right)
	default:
		return cell + strings.Repeat(" ", pad)
	}
}
