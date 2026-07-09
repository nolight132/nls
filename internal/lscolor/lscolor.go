package lscolor

import (
	"os"
	"strings"

	"github.com/nolight132/nls/internal/listing"
)

type suffixRule struct {
	sequence string
	// index preserves dircolors semantics: of all rules matching a name,
	// the one latest in LS_COLORS wins.
	index int
}

type otherRule struct {
	suffix   string
	sequence string
	index    int
}

// Styler applies LS_COLORS-compatible filename colors.
type Styler struct {
	// dotRules holds "*.ext" selectors keyed by ".ext"; otherRules holds
	// the rare "*suffix" selectors whose suffix does not start with a dot.
	dotRules   map[string]suffixRule
	otherRules []otherRule
	typeRules  map[string]string
}

// New reads LS_COLORS or falls back to a small theme-friendly default.
func New() *Styler {
	raw := os.Getenv("LS_COLORS")
	if raw == "" {
		raw = defaultColors
	}
	return parse(raw)
}

func parse(raw string) *Styler {
	s := &Styler{dotRules: make(map[string]suffixRule), typeRules: make(map[string]string)}
	index := 0
	for entry := range strings.SplitSeq(raw, ":") {
		if entry == "" {
			continue
		}
		selector, sequence, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(selector, "*"):
			suffix := selector[1:]
			if strings.HasPrefix(suffix, ".") {
				s.dotRules[suffix] = suffixRule{sequence: sequence, index: index}
			} else {
				s.otherRules = append(s.otherRules, otherRule{suffix: suffix, sequence: sequence, index: index})
			}
			index++
		case strings.Contains(selector, "*"):
			// Mid-string wildcards were never matched; keep ignoring them.
		default:
			s.typeRules[selector] = sequence
		}
	}
	return s
}

// Colorize wraps a filename using LS_COLORS rules.
func (s *Styler) Colorize(name string, kind listing.Kind) string {
	seq := s.matchSequence(name, kind)
	if seq == "" || seq == "0" || !validSequence(seq) {
		return name
	}
	return "\x1b[" + seq + "m" + name + "\x1b[0m"
}

// validSequence rejects values that are not SGR parameters, like the
// dircolors keyword ln=target, so they never reach the terminal raw.
func validSequence(seq string) bool {
	for i := 0; i < len(seq); i++ {
		c := seq[i]
		if (c < '0' || c > '9') && c != ';' {
			return false
		}
	}
	return true
}

func (s *Styler) matchSequence(name string, kind listing.Kind) string {
	seq, best := "", -1
	for i := 0; i < len(name); i++ {
		if name[i] != '.' {
			continue
		}
		if r, ok := s.dotRules[name[i:]]; ok && r.index > best {
			seq, best = r.sequence, r.index
		}
	}
	for _, r := range s.otherRules {
		if r.index > best && strings.HasSuffix(name, r.suffix) {
			seq, best = r.sequence, r.index
		}
	}
	if seq != "" {
		return seq
	}
	return s.typeRules[typeCode(kind)]
}

func typeCode(kind listing.Kind) string {
	switch kind {
	case listing.KindDirectory:
		return "di"
	case listing.KindSymlink:
		return "ln"
	case listing.KindExecutable:
		return "ex"
	default:
		return "fi"
	}
}
