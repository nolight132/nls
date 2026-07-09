package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
