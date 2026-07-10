package cli

import (
	"os"

	"github.com/nolight132/nls/internal/config"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/output"
	"golang.org/x/term"
)

func buildListOptions(cfg *Flags, userCfg config.Config, interactive bool) listing.ListOptions {
	estimateSizes := cfg.Precise || (interactive && userCfg.DirSize.Enabled)
	estimateDepth := 0
	if cfg.Precise {
		estimateDepth = listing.EstimateDepthMax
	} else if estimateSizes {
		estimateDepth = listing.EstimateDepthBounded
	}

	// Git status is computed when the -g column asks for it, or when
	// default git coloring needs the data on a colored interactive
	// listing. The column itself still only renders with -g.
	gitStatus := cfg.GitStatus ||
		(interactive && !cfg.NoColor && userCfg.Git.ColorEntries)

	needLinkTarget := cfg.JSON || cfg.Long || userCfg.Render.ShowLinkTarget
	needLinkTargetDir := cfg.DirsFirst || (needLinkTarget && cfg.Classify)

	return listing.ListOptions{
		DirSizeDepth:      userCfg.DirSize.DefaultDepth,
		DirSizeTiming:     userCfg.DirSize.Timing,
		All:               cfg.All,
		AlmostAll:         cfg.AlmostAll,
		IgnoreBackups:     cfg.IgnoreBack,
		Dereference:       cfg.Dereference,
		Directory:         cfg.Directory,
		Recursive:         cfg.Recursive,
		EstimateSizes:     estimateSizes,
		EstimateDepth:     estimateDepth,
		Precise:           cfg.Precise,
		LongListing:       cfg.Long,
		ShowInode:         cfg.Inode,
		ShowBlocks:        cfg.Blocks,
		Classify:          cfg.Classify,
		DirSlash:          cfg.DirSlash,
		QuoteNames:        cfg.QuoteName,
		Commas:            cfg.Commas,
		Sort:              buildSort(cfg),
		GitStatus:         gitStatus,
		NeedLinkTarget:    needLinkTarget,
		NeedLinkTargetDir: needLinkTargetDir,
	}
}

func buildColumns(flags *Flags, userCfg config.Config) []string {
	type optional struct {
		column config.ColumnEntry
		flag   bool
	}
	opt := []optional{
		{column: config.ColumnInode, flag: flags.Inode},
		{column: config.ColumnBlocks, flag: flags.Blocks},
		{column: config.ColumnPermissions, flag: flags.Long},
		{column: config.ColumnOwner, flag: flags.Long},
		{column: config.ColumnGitStatus, flag: flags.GitStatus},
	}
	cols := make([]string, 0, len(userCfg.DefaultColumns)+len(opt))
	seen := make(map[config.ColumnEntry]bool, len(userCfg.DefaultColumns)+len(opt))
	for _, c := range userCfg.DefaultColumns {
		s := string(c)
		if s == string(config.ColumnModified) {
			s = timeColumn(flags)
		}
		if !seen[config.ColumnEntry(s)] {
			cols = append(cols, s)
			seen[config.ColumnEntry(s)] = true
		}
	}
	for _, o := range opt {
		if o.flag && !seen[o.column] {
			cols = append(cols, string(o.column))
			seen[o.column] = true
		}
	}

	return cols
}

func buildSort(cfg *Flags) listing.SortOptions {
	sort := listing.SortOptions{
		Reverse:   cfg.Reverse,
		DirsFirst: cfg.DirsFirst,
		TimeField: timeField(cfg),
	}
	switch {
	case cfg.Unsorted:
		sort.Field = listing.SortByNone
	case cfg.SortTime || ((cfg.SortAccess || cfg.SortChange) && !cfg.Long):
		sort.Field = listing.SortByTime
	case cfg.SortSize:
		sort.Field = listing.SortBySize
	case cfg.SortExt:
		sort.Field = listing.SortByExtension
	default:
		sort.Field = listing.SortByName
	}
	return sort
}

func timeField(cfg *Flags) listing.TimeField {
	switch {
	case cfg.SortAccess:
		return listing.TimeAccessed
	case cfg.SortChange:
		return listing.TimeChanged
	default:
		return listing.TimeModified
	}
}

// timeColumn maps -u/-c to the column showing the sorted timestamp.
func timeColumn(cfg *Flags) string {
	switch timeField(cfg) {
	case listing.TimeAccessed:
		return string(config.ColumnAccessed)
	case listing.TimeChanged:
		return string(config.ColumnChanged)
	default:
		return string(config.ColumnModified)
	}
}

func plainMode(cfg *Flags) output.PlainMode {
	switch {
	case cfg.Commas:
		return output.PlainCommas
	case cfg.Long:
		return output.PlainLong
	default:
		return output.PlainOne
	}
}

// useTable is true on a TTY unless the user asks for a different output shape.
func useTable(cfg *Flags, isTTY bool) bool {
	if cfg.Table {
		return true
	}
	if !isTTY || cfg.JSON || cfg.Plain {
		return false
	}
	if cfg.Commas || cfg.One {
		return false
	}
	return true
}

// useColor is true on a TTY unless colors are disabled explicitly.
func useColor(cfg *Flags, isTTY bool) bool {
	if cfg.NoColor {
		return false
	}
	return isTTY
}

// terminalWidth returns the stdout terminal width for table capping,
// or 0 when not interactive or the size cannot be determined.
func terminalWidth(interactive bool) int {
	if !interactive {
		return 0
	}
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 0
	}
	return width
}
