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
	header       string
	align        cellAlign
	centerHeader bool
	render       func(e listing.Entry, idx int, ctx renderCtx) string
}

type renderCtx struct {
	opts   Options
	styles *termcolor.Style
	now    time.Time
	human  bool
}

var columnRegistry = map[string]struct {
	header       string
	align        cellAlign
	centerHeader bool
	render       func(e listing.Entry, idx int, ctx renderCtx) string
}{
	"id": {
		header: "#", align: alignRight, centerHeader: true,
		render: func(e listing.Entry, idx int, ctx renderCtx) string {
			return ctx.styles.Index(strconv.Itoa(idx))
		},
	},
	"name": {
		header: "name", align: alignLeft, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Name(tableDisplayName(e, ctx.opts), e.Kind)
		},
	},
	"type": {
		header: "type", align: alignLeft, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return typeLabel(e)
		},
	},
	"size": {
		header: "size", align: alignRight, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Size(format.Size(e.Size, ctx.human, e.SizeApprox))
		},
	},
	"modified": {
		header: "modified", align: alignLeft, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Modified(tableTimeField(e.Modified, ctx.opts, ctx.now))
		},
	},
	"accessed": {
		header: "accessed", align: alignLeft, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Modified(tableTimeField(e.Accessed, ctx.opts, ctx.now))
		},
	},
	"changed": {
		header: "changed", align: alignLeft, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Modified(tableTimeField(e.Changed, ctx.opts, ctx.now))
		},
	},
	"permissions": {
		header: "permissions", align: alignLeft, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return e.Permissions
		},
	},
	"links": {
		header: "links", align: alignRight, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return strconv.FormatUint(e.Links, 10)
		},
	},
	"owner": {
		header: "owner", align: alignLeft, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return e.Owner
		},
	},
	"group": {
		header: "group", align: alignLeft, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return e.Group
		},
	},
	"inode": {
		header: "inode", align: alignRight, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return strconv.FormatUint(e.Inode, 10)
		},
	},
	"blocks": {
		header: "blocks", align: alignRight, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return strconv.FormatInt(e.Blocks, 10)
		},
	},
}

var defaultTableColumns = []string{"id", "name", "type", "size", "modified"}

func buildTableColumns(opts Options, styles *termcolor.Style) []tableColumn {
	names := opts.Columns
	if len(names) == 0 {
		names = defaultTableColumns
	}
	cols := make([]tableColumn, 0, len(names))
	for _, name := range names {
		spec, ok := columnRegistry[name]
		if !ok {
			continue
		}
		cols = append(cols, tableColumn{
			header:       styles.Header(spec.header),
			align:        spec.align,
			centerHeader: spec.centerHeader,
			render:       spec.render,
		})
	}
	return cols
}

func renderTable(w io.Writer, entries []listing.Entry, opts Options) error {
	styles := termcolor.New(opts.Color)
	cols := buildTableColumns(opts, styles)
	if len(cols) == 0 {
		return nil
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
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

	table := buildBorderedTable(cols, rows)
	_, err := fmt.Fprint(w, table)
	return err
}

func tableDisplayName(e listing.Entry, opts Options) string {
	name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName, true)
	return icons.For(e, opts.IconSet) + name
}

func tableTimeField(t time.Time, opts Options, now time.Time) string {
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
		if col.centerHeader {
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
