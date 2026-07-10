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
	// flex marks the column that shrinks when the table exceeds the
	// terminal width.
	flex        bool
	width       int
	headerWidth int
	// subDivider is the display offset of a vertical rule inside the
	// column's cells (the git status separator); -1 when the column has
	// none. Horizontal borders hook into it with ┬ and ┴.
	subDivider int
	render     func(e listing.Entry, idx int, ctx renderCtx) string
}

type renderCtx struct {
	opts   RenderOptions
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
			name := tableDisplayName(e, ctx.opts)
			if ctx.opts.GitColorEntries {
				return ctx.styles.NameGit(name, e.Kind, e.GitState)
			}
			return ctx.styles.Name(name, e.Kind)
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
			return ctx.styles.Modified(format.Modified(e.Modified, ctx.now))
		},
	},
	"accessed": {
		header: "accessed", align: alignLeft, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Modified(format.Modified(e.Accessed, ctx.now))
		},
	},
	"changed": {
		header: "changed", align: alignLeft, centerHeader: true,
		render: func(e listing.Entry, _ int, ctx renderCtx) string {
			return ctx.styles.Modified(format.Modified(e.Changed, ctx.now))
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
	"git": {
		header: "git", align: alignLeft, centerHeader: false,
		render: func(e listing.Entry, _ int, _ renderCtx) string {
			return string(rune(e.GitState.Staging)) + "│" + string(rune(e.GitState.Worktree))
		},
	},
}

func buildTableColumns(opts RenderOptions, styles *termcolor.Style) []tableColumn {
	names := opts.Columns
	cols := make([]tableColumn, 0, len(names))
	for _, name := range names {
		spec, ok := columnRegistry[name]
		if !ok {
			continue
		}
		subDivider := -1
		if name == "git" {
			subDivider = 1
		}
		header := styles.Header(spec.header)
		cols = append(cols, tableColumn{
			header:       header,
			align:        spec.align,
			centerHeader: spec.centerHeader,
			flex:         name == "name",
			subDivider:   subDivider,
			headerWidth:  visibleWidth(header),
			render:       spec.render,
		})
	}
	return cols
}

func renderTable(w io.Writer, entries []listing.Entry, opts RenderOptions, styles *termcolor.Style) error {
	if len(entries) == 0 {
		renderEmptyTable(w, styles)
		return nil
	}

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

	table := buildBorderedTable(cols, rows, opts.Width)
	_, err := fmt.Fprint(w, table)
	return err
}

func renderEmptyTable(w io.Writer, styles *termcolor.Style) {
	var b strings.Builder
	emptyCols := []tableColumn{{header: "", width: 10, subDivider: -1}}
	emptyMessage := styles.Empty("no entries")
	writeBorderTop(&b, emptyCols)
	writeDataRow(&b, emptyCols, []string{emptyMessage}, []int{visibleWidth(emptyMessage)})
	writeBorderBottom(&b, emptyCols)
	fmt.Fprint(w, b.String())
}

func tableDisplayName(e listing.Entry, opts RenderOptions) string {
	showLinkTarget := opts.ShowLinkTarget
	if opts.Long {
		showLinkTarget = true
	}
	name := listing.DisplayName(e, opts.Classify, opts.DirSlash, opts.QuoteName, showLinkTarget)
	if opts.IsTTY || opts.UseTable {
		// Tables are presentation-only, so control characters are hidden
		// even when piped; plain piped output keeps raw names for scripts.
		name = sanitizeName(name)
	}
	return icons.For(e, opts.IconSet) + name
}

func buildBorderedTable(cols []tableColumn, rows [][]string, limit int) string {
	widths := measureRows(rows)
	computeWidths(cols, widths)
	fitWidths(cols, rows, widths, limit)

	var b strings.Builder
	writeBorderTop(&b, cols)
	writeHeaderRow(&b, cols)
	writeBorderMid(&b, cols)
	for ri, row := range rows {
		writeDataRow(&b, cols, row, widths[ri])
	}
	writeBorderBottom(&b, cols)
	return b.String()
}

// measureRows caches every cell's display width so alignment and fitting
// never re-strip ANSI from the same string.
func measureRows(rows [][]string) [][]int {
	widths := make([][]int, len(rows))
	for ri, row := range rows {
		rw := make([]int, len(row))
		for ci, cell := range row {
			rw[ci] = visibleWidth(cell)
		}
		widths[ri] = rw
	}
	return widths
}

func computeWidths(cols []tableColumn, widths [][]int) {
	for i := range cols {
		cols[i].width = cols[i].headerWidth
	}
	for _, rw := range widths {
		for i, w := range rw {
			if w > cols[i].width {
				cols[i].width = w
			}
		}
	}
}

// minFlexWidth keeps a shrunken name column readable.
const minFlexWidth = 8

// fitWidths shrinks the flex column and truncates its cells so the
// rendered table fits within limit display cells.
func fitWidths(cols []tableColumn, rows [][]string, widths [][]int, limit int) {
	if limit <= 0 {
		return
	}
	total := tableWidth(cols)
	if total <= limit {
		return
	}
	flexIdx := -1
	for i := range cols {
		if cols[i].flex {
			flexIdx = i
			break
		}
	}
	if flexIdx == -1 {
		return
	}
	width := cols[flexIdx].width - (total - limit)
	if min := max(minFlexWidth, cols[flexIdx].headerWidth); width < min {
		width = min
	}
	if width >= cols[flexIdx].width {
		return
	}
	cols[flexIdx].width = width
	for ri, row := range rows {
		if widths[ri][flexIdx] <= width {
			continue
		}
		row[flexIdx] = truncateANSI(row[flexIdx], width)
		widths[ri][flexIdx] = visibleWidth(row[flexIdx])
	}
}

func tableWidth(cols []tableColumn) int {
	total := 1
	for _, col := range cols {
		total += col.width + 3
	}
	return total
}

func writeBorderTop(b *strings.Builder, cols []tableColumn) {
	b.WriteString("╭")
	for i, col := range cols {
		if i > 0 {
			b.WriteString("┬")
		}
		b.WriteString(strings.Repeat("─", col.width+2))
	}
	b.WriteString("╮\n")
}

func writeBorderMid(b *strings.Builder, cols []tableColumn) {
	b.WriteString("├")
	for i, col := range cols {
		if i > 0 {
			b.WriteString("┼")
		}
		// ┬ starts the sub-divider here rather than in the top border so
		// the header text above it stays unbroken.
		b.WriteString(borderSegment(col, "┬"))
	}
	b.WriteString("┤\n")
}

func writeBorderBottom(b *strings.Builder, cols []tableColumn) {
	b.WriteString("╰")
	for i, col := range cols {
		if i > 0 {
			b.WriteString("┴")
		}
		b.WriteString(borderSegment(col, "┴"))
	}
	b.WriteString("╯\n")
}

// borderSegment draws a column's stretch of a horizontal border, hooking
// in the cell-internal divider with the given junction when present.
func borderSegment(col tableColumn, junction string) string {
	at := 1 + col.subDivider
	if col.subDivider < 0 || at >= col.width+2 {
		return strings.Repeat("─", col.width+2)
	}
	return strings.Repeat("─", at) + junction + strings.Repeat("─", col.width+1-at)
}

func writeHeaderRow(b *strings.Builder, cols []tableColumn) {
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
		b.WriteString(alignCell(col.header, col.headerWidth, col.width, align))
		b.WriteString(" ")
	}
	b.WriteString("│\n")
}

func writeDataRow(b *strings.Builder, cols []tableColumn, row []string, widths []int) {
	b.WriteRune('│')
	for i, cell := range row {
		if i > 0 {
			b.WriteRune('│')
		}
		b.WriteString(" ")
		b.WriteString(alignCell(cell, widths[i], cols[i].width, cols[i].align))
		b.WriteString(" ")
	}
	b.WriteString("│\n")
}

func alignCell(cell string, cellWidth, width int, align cellAlign) string {
	pad := width - cellWidth
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
