package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/nolight132/nls/internal/config"
	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/output"
	"github.com/nolight132/nls/internal/pathutil"
	"github.com/nolight132/nls/internal/version"
	"github.com/spf13/cobra"
)

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
	Plain       bool
	Table       bool
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
	cmd.Flags().BoolVar(&cfg.Plain, "plain", false, "output in plain text")
	cmd.Flags().BoolVar(&cfg.Table, "table", false, "output in table format")
	cmd.Flags().BoolP("version", "", false, "version for nls")
	configureHelp(cmd)

	return cmd
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
	colorEnabled := useColor(cfg, isTTY)

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
		ShowLinkTarget:  userCfg.Render.ShowLinkTarget,
	}

	blocks, errs := listing.List(expanded, listOpts)
	suggest := output.StderrIsTTY()
	for _, e := range errs {
		output.WriteError(e, suggest)
	}
	out := bufio.NewWriter(os.Stdout)
	if err := output.Render(out, blocks, outOpts); err != nil {
		return err
	}
	return out.Flush()
}

func loadUserConfig(w io.Writer) config.Config {
	userCfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(w, "nls: warning: %v; using defaults\n", err)
		return config.Defaults()
	}
	return userCfg
}
