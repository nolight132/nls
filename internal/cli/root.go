package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nolight132/nls/internal/config"
	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/output"
	"github.com/nolight132/nls/internal/pathutil"
	"github.com/nolight132/nls/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// ErrReported signals a nonzero exit for errors already written to stderr.
var ErrReported = errors.New("errors already reported")

// Flags holds parsed CLI flags.
type Flags struct {
	All         bool
	AlmostAll   bool
	Long        bool
	Human       bool
	One         bool
	Recursive   bool
	Reverse     bool
	SortTime    bool
	SortSize    bool
	SortExt     bool
	Unsorted    bool
	Directory   bool
	Classify    bool
	DirSlash    bool
	IgnoreBack  bool
	Dereference bool
	Commas      bool
	QuoteName   bool
	DirsFirst   bool
	Inode       bool
	Blocks      bool
	SortAccess  bool
	SortChange  bool
	NoIcons     bool
	NoColor     bool
	JSON        bool
	Precise     bool
	Paths       []string
	GitStatus   bool
}

// Root returns the root cobra command.
func Root() *cobra.Command {
	cfg := &Flags{}
	var unsortedF bool

	cmd := &cobra.Command{
		Use:           "nls [path...]",
		Short:         "List directory contents",
		Long:          "nls lists files and directories with Nushell-style columns for use in non-Nu shells.",
		Version:       version.String(),
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Paths = args
			if unsortedF {
				// POSIX -f: unsorted and all entries.
				cfg.Unsorted = true
				cfg.All = true
			}
			return run(cfg)
		},
	}

	cmd.Flags().Bool("help", false, "help for nls")
	cmd.Flags().BoolVarP(&cfg.All, "all", "a", false, "do not hide entries starting with .")
	cmd.Flags().BoolVarP(&cfg.AlmostAll, "almost-all", "A", false, "like -a but omit . and ..")
	cmd.Flags().BoolVarP(&cfg.Long, "long", "l", false, "show extended metadata")
	cmd.Flags().BoolVarP(&cfg.Human, "human-readable", "h", false, "print human-readable sizes")
	cmd.Flags().BoolVarP(&cfg.One, "one", "1", false, "list one file per line")
	cmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "R", false, "list subdirectories recursively")
	cmd.Flags().BoolVarP(&cfg.Reverse, "reverse", "r", false, "reverse order while sorting")
	cmd.Flags().BoolVarP(&cfg.SortTime, "time", "t", false, "sort by modification time")
	cmd.Flags().BoolVarP(&cfg.SortSize, "size", "S", false, "sort by file size")
	cmd.Flags().BoolVarP(&cfg.SortExt, "extension", "X", false, "sort alphabetically by extension")
	cmd.Flags().BoolVarP(&cfg.Unsorted, "unsorted", "U", false, "do not sort")
	cmd.Flags().BoolVarP(&unsortedF, "fast", "f", false, "do not sort, list all entries (as -aU)")
	cmd.Flags().BoolVarP(&cfg.Directory, "directory", "d", false, "list directories themselves, not contents")
	cmd.Flags().BoolVarP(&cfg.Classify, "classify", "F", false, "append indicator (one of */=>@|)")
	cmd.Flags().BoolVarP(&cfg.DirSlash, "slash", "p", false, "append / to directory names")
	cmd.Flags().BoolVarP(&cfg.IgnoreBack, "ignore-backups", "B", false, "do not list implied entries ending with ~")
	cmd.Flags().BoolVarP(&cfg.Dereference, "dereference", "L", false, "follow symlinks")
	cmd.Flags().BoolVarP(&cfg.Commas, "comma", "m", false, "fill width with a comma separated list")
	cmd.Flags().BoolVarP(&cfg.QuoteName, "quote-name", "Q", false, "enclose entry names in double quotes")
	cmd.Flags().BoolVar(&cfg.DirsFirst, "group-directories-first", false, "group directories before files")
	cmd.Flags().BoolVarP(&cfg.Inode, "inode", "i", false, "show inode numbers")
	cmd.Flags().BoolVarP(&cfg.Blocks, "size-blocks", "s", false, "show allocated block counts")
	cmd.Flags().BoolVarP(&cfg.SortAccess, "access-time", "u", false, "sort by access time")
	cmd.Flags().BoolVarP(&cfg.SortChange, "ctime", "c", false, "sort by status change time")
	cmd.Flags().BoolVar(&cfg.NoIcons, "no-icons", false, "disable icons")
	cmd.Flags().BoolVar(&cfg.NoColor, "no-color", false, "disable colors")
	cmd.Flags().BoolVar(&cfg.JSON, "json", false, "output JSON")
	cmd.Flags().BoolVarP(&cfg.Precise, "precise", "P", false, "compute exact directory sizes without depth, time, or entry limits")
	cmd.Flags().BoolVarP(&cfg.GitStatus, "git-status", "g", false, "show git status")
	cmd.Flags().BoolP("version", "", false, "version for nls")
	configureHelp(cmd)

	return cmd
}

func configureHelp(cmd *cobra.Command) {
	markGroup(cmd, "Listing flags (table and plain output)",
		"all", "almost-all", "long", "human-readable", "recursive", "reverse",
		"time", "access-time", "ctime", "size", "extension", "unsorted", "fast",
		"directory", "classify", "slash", "ignore-backups", "dereference",
		"quote-name", "group-directories-first", "inode", "size-blocks",
		"git-status",
	)
	markGroup(cmd, "Plain-output layout flags", "one", "comma")
	markGroup(cmd, "nls presentation flags", "json", "precise", "no-icons", "no-color", "help", "version")

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "%s\n\n", cmd.Long)
		fmt.Fprintf(out, "Usage:\n  %s\n\n", cmd.UseLine())

		writeFlagGroup(out, cmd.Flags(), "Listing flags (table and plain output)")
		writeFlagGroup(out, cmd.Flags(), "Plain-output layout flags")
		writeFlagGroup(out, cmd.Flags(), "nls presentation flags")
	})
}

func markGroup(cmd *cobra.Command, group string, names ...string) {
	for _, name := range names {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			continue
		}
		if flag.Annotations == nil {
			flag.Annotations = map[string][]string{}
		}
		flag.Annotations["nls:group"] = []string{group}
	}
}

func writeFlagGroup(out io.Writer, flags *pflag.FlagSet, group string) {
	var lines []string
	flags.VisitAll(func(flag *pflag.Flag) {
		flagGroup, ok := flag.Annotations["nls:group"]
		if flag.Hidden || !ok || len(flagGroup) == 0 || flagGroup[0] != group {
			return
		}
		lines = append(lines, formatFlag(flag))
	})
	if len(lines) == 0 {
		return
	}
	fmt.Fprintf(out, "%s:\n%s\n\n", group, strings.Join(lines, "\n"))
}

func formatFlag(flag *pflag.Flag) string {
	name := "      --" + flag.Name
	if flag.Shorthand != "" {
		name = fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
	}

	usage := strings.TrimSpace(flag.Usage)
	if usage == "" {
		return fmt.Sprintf("%-34s", name)
	}

	lines := strings.Split(usage, "\n")
	const nameWidth = 34
	var b strings.Builder
	fmt.Fprintf(&b, "%-*s %s", nameWidth, name, strings.TrimSpace(lines[0]))
	indent := strings.Repeat(" ", nameWidth+1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteByte('\n')
		b.WriteString(indent)
		b.WriteString(line)
	}
	return b.String()
}

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

	return listing.ListOptions{
		DirSizeDepth:  userCfg.DirSize.DefaultDepth,
		DirSizeTiming: userCfg.DirSize.Timing,
		All:           cfg.All,
		AlmostAll:     cfg.AlmostAll,
		IgnoreBackups: cfg.IgnoreBack,
		Dereference:   cfg.Dereference,
		Directory:     cfg.Directory,
		Recursive:     cfg.Recursive,
		EstimateSizes: estimateSizes,
		EstimateDepth: estimateDepth,
		Precise:       cfg.Precise,
		LongListing:   cfg.Long,
		ShowInode:     cfg.Inode,
		ShowBlocks:    cfg.Blocks,
		Classify:      cfg.Classify,
		DirSlash:      cfg.DirSlash,
		QuoteNames:    cfg.QuoteName,
		Commas:        cfg.Commas,
		Sort:          buildSort(cfg),
		GitStatus:     gitStatus,
	}
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

func run(cfg *Flags) error {
	userCfg := loadUserConfig(os.Stderr)

	paths := cfg.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	expanded := make([]string, 0, len(paths))
	for _, raw := range paths {
		p, err := pathutil.Expand(raw)
		if err != nil {
			return err
		}
		expanded = append(expanded, p)
	}

	isTTY := output.StdoutIsTTY()
	interactive := useTable(cfg, isTTY)
	colorEnabled := interactive && !cfg.NoColor

	var iconSet icons.Set
	if interactive {
		iconSet = icons.Resolve(cfg.NoIcons, userCfg.Icons.Enabled, userCfg.Icons.SpecialIcons)
	}

	listOpts := buildListOptions(cfg, userCfg, interactive)

	outOpts := output.RenderOptions{
		Human:           cfg.Human || interactive,
		Long:            cfg.Long,
		JSON:            cfg.JSON,
		Color:           colorEnabled,
		IconSet:         iconSet,
		IsTTY:           isTTY,
		Plain:           plainMode(cfg),
		Classify:        cfg.Classify,
		DirSlash:        cfg.DirSlash,
		QuoteName:       cfg.QuoteName,
		ShowInode:       cfg.Inode,
		ShowBlocks:      cfg.Blocks,
		UseTable:        interactive,
		Width:           terminalWidth(interactive),
		Columns:         buildColumns(cfg, userCfg),
		GitColorEntries: userCfg.Git.ColorEntries,
	}

	blocks, errs := listing.List(expanded, listOpts)
	suggest := output.StderrIsTTY()
	for _, e := range errs {
		output.WriteError(e, suggest)
	}
	if err := output.Render(os.Stdout, blocks, outOpts); err != nil {
		return err
	}
	if len(errs) > 0 {
		return ErrReported
	}
	return nil
}

func loadUserConfig(w io.Writer) config.Config {
	userCfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(w, "nls: warning: %v; using defaults\n", err)
		return config.Defaults()
	}
	return userCfg
}

func buildColumns(cfg *Flags, userCfg config.Config) []string {
	cols := make([]string, 0, len(userCfg.DefaultColumns)+4)
	seen := make(map[string]bool, len(userCfg.DefaultColumns)+4)
	for _, c := range userCfg.DefaultColumns {
		s := string(c)
		if s == string(config.ColumnModified) {
			s = timeColumn(cfg)
		}
		if !seen[s] {
			cols = append(cols, s)
			seen[s] = true
		}
	}
	if cfg.Inode && !seen[string(config.ColumnInode)] {
		cols = append(cols, string(config.ColumnInode))
	}
	if cfg.Blocks && !seen[string(config.ColumnBlocks)] {
		cols = append(cols, string(config.ColumnBlocks))
	}
	if cfg.Long && !seen[string(config.ColumnPermissions)] {
		cols = append(cols, string(config.ColumnPermissions))
	}
	if cfg.GitStatus && !seen[string(config.ColumnGitStatus)] {
		cols = append(cols, string(config.ColumnGitStatus))
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
	if !isTTY || cfg.JSON {
		return false
	}
	if cfg.Commas || cfg.One {
		return false
	}
	return true
}
