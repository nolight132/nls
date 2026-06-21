package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/output"
	"github.com/nolight132/nls/internal/pathutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Config holds parsed CLI flags.
type Config struct {
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
	FullTime    bool
	DirsFirst   bool
	Inode       bool
	Blocks      bool
	SortAccess  bool
	SortChange  bool
	NoIcons     bool
	NoColor     bool
	JSON        bool
	Paths       []string
}

// Root returns the root cobra command.
func Root() *cobra.Command {
	cfg := &Config{}
	var unsortedF bool

	cmd := &cobra.Command{
		Use:           "nls [path...]",
		Short:         "List directory contents",
		Long:          "nls lists files and directories with Nushell-style columns for use in non-Nu shells.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Paths = args
			if unsortedF {
				cfg.Unsorted = true
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
	cmd.Flags().BoolVarP(&unsortedF, "fast", "f", false, "do not sort (as -U)")
	cmd.Flags().BoolVarP(&cfg.Directory, "directory", "d", false, "list directories themselves, not contents")
	cmd.Flags().BoolVarP(&cfg.Classify, "classify", "F", false, "append indicator (one of */=>@|)")
	cmd.Flags().BoolVarP(&cfg.DirSlash, "slash", "p", false, "append / to directory names")
	cmd.Flags().BoolVarP(&cfg.IgnoreBack, "ignore-backups", "B", false, "do not list implied entries ending with ~")
	cmd.Flags().BoolVarP(&cfg.Dereference, "dereference", "L", false, "follow symlinks")
	cmd.Flags().BoolVarP(&cfg.Commas, "comma", "m", false, "fill width with a comma separated list")
	cmd.Flags().BoolVarP(&cfg.QuoteName, "quote-name", "Q", false, "enclose entry names in double quotes")
	cmd.Flags().BoolVar(&cfg.FullTime, "full-time", false, "print full timestamps")
	cmd.Flags().BoolVar(&cfg.DirsFirst, "group-directories-first", false, "group directories before files")
	cmd.Flags().BoolVarP(&cfg.Inode, "inode", "i", false, "show inode numbers")
	cmd.Flags().BoolVarP(&cfg.Blocks, "size-blocks", "s", false, "show allocated block counts")
	cmd.Flags().BoolVarP(&cfg.SortAccess, "access-time", "u", false, "sort by access time")
	cmd.Flags().BoolVarP(&cfg.SortChange, "ctime", "c", false, "sort by status change time")
	cmd.Flags().BoolVar(&cfg.NoIcons, "no-icons", false, "disable icons")
	cmd.Flags().BoolVar(&cfg.NoColor, "no-color", false, "disable colors")
	cmd.Flags().BoolVar(&cfg.JSON, "json", false, "output JSON")
	configureHelp(cmd)

	return cmd
}

func configureHelp(cmd *cobra.Command) {
	markGroup(cmd, "Listing flags (table and plain output)",
		"all", "almost-all", "long", "human-readable", "recursive", "reverse",
		"time", "access-time", "ctime", "size", "extension", "unsorted", "fast",
		"directory", "classify", "slash", "ignore-backups", "dereference",
		"quote-name", "full-time", "group-directories-first", "inode", "size-blocks",
	)
	markGroup(cmd, "Plain-output layout flags", "one", "comma")
	markGroup(cmd, "nls presentation flags", "json", "no-icons", "no-color", "help")

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
	return fmt.Sprintf("%-34s %s", name, flag.Usage)
}

func buildListOptions(cfg *Config, interactive bool) listing.Options {
	needsFull := listing.NeedsFullMetadata(listing.Options{
		All:              cfg.All,
		AlmostAll:        cfg.AlmostAll,
		Dereference:      cfg.Dereference,
		Directory:        cfg.Directory,
		Recursive:        cfg.Recursive,
		EstimateDirSizes: interactive,
		LongListing:      cfg.Long,
		ShowInode:        cfg.Inode,
		ShowBlocks:       cfg.Blocks,
		Classify:         cfg.Classify,
		DirSlash:         cfg.DirSlash,
		Sort:             buildSort(cfg),
	})

	return listing.Options{
		All:              cfg.All,
		AlmostAll:        cfg.AlmostAll,
		IgnoreBackups:    cfg.IgnoreBack,
		Dereference:      cfg.Dereference,
		Directory:        cfg.Directory,
		Recursive:        cfg.Recursive,
		EstimateDirSizes: interactive,
		FastPath:         !interactive && !cfg.JSON && !needsFull,
		ResolveAbs:       interactive || cfg.JSON,
		LongListing:      cfg.Long,
		ShowInode:        cfg.Inode,
		ShowBlocks:       cfg.Blocks,
		Classify:         cfg.Classify,
		DirSlash:         cfg.DirSlash,
		QuoteNames:       cfg.QuoteName,
		Commas:           cfg.Commas,
		Sort:             buildSort(cfg),
	}
}

func run(cfg *Config) error {
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
		iconSet = icons.Resolve(cfg.NoIcons)
	}

	listOpts := buildListOptions(cfg, interactive)

	outOpts := output.Options{
		Human:      cfg.Human || interactive,
		Long:       cfg.Long,
		JSON:       cfg.JSON,
		Color:      colorEnabled,
		IconSet:    iconSet,
		IsTTY:      isTTY,
		Plain:      plainMode(cfg, isTTY),
		Classify:   cfg.Classify,
		DirSlash:   cfg.DirSlash,
		QuoteName:  cfg.QuoteName,
		FullTime:   cfg.FullTime,
		ShowInode:  cfg.Inode,
		ShowBlocks: cfg.Blocks,
		TimeField:  timeField(cfg),
		UseTable:   interactive,
	}

	return output.RenderFast(os.Stdout, expanded, listOpts, outOpts)
}

func buildSort(cfg *Config) listing.SortOptions {
	sort := listing.SortOptions{
		Reverse:   cfg.Reverse,
		DirsFirst: cfg.DirsFirst,
		TimeField: timeField(cfg),
	}
	switch {
	case cfg.Unsorted:
		sort.Field = listing.SortByNone
	case cfg.SortTime || cfg.SortAccess || cfg.SortChange:
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

func timeField(cfg *Config) listing.TimeField {
	switch {
	case cfg.SortAccess:
		return listing.TimeAccessed
	case cfg.SortChange:
		return listing.TimeChanged
	default:
		return listing.TimeModified
	}
}

func plainMode(cfg *Config, isTTY bool) output.PlainMode {
	if cfg.Commas {
		return output.PlainCommas
	}
	if cfg.Long {
		return output.PlainLong
	}
	if cfg.One || !isTTY {
		return output.PlainOne
	}
	return output.PlainOne
}

// useTable is true on a TTY unless the user asks for a different output shape.
func useTable(cfg *Config, isTTY bool) bool {
	if !isTTY || cfg.JSON {
		return false
	}
	if cfg.Commas || cfg.One {
		return false
	}
	return true
}
