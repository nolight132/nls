package lscolor

import (
	"os"
	"strings"

	"github.com/nolight132/nls/internal/listing"
)

type rule struct {
	selector string
	sequence string
}

// Styler applies LS_COLORS-compatible filename colors.
type Styler struct {
	extRules  []rule
	typeRules map[string]string
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
	s := &Styler{typeRules: make(map[string]string)}
	for entry := range strings.SplitSeq(raw, ":") {
		if entry == "" {
			continue
		}
		selector, sequence, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if strings.Contains(selector, "*") {
			s.extRules = append(s.extRules, rule{selector: selector, sequence: sequence})
			continue
		}
		s.typeRules[selector] = sequence
	}
	return s
}

// Colorize wraps a filename using LS_COLORS rules.
func (s *Styler) Colorize(name string, kind listing.Kind) string {
	seq := s.matchSequence(name, kind)
	if seq == "" || seq == "0" {
		return name
	}
	return "\x1b[" + seq + "m" + name + "\x1b[0m"
}

func (s *Styler) matchSequence(name string, kind listing.Kind) string {
	seq := ""
	for _, r := range s.extRules {
		if matchSelector(r.selector, name) {
			seq = r.sequence
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

func matchSelector(selector, name string) bool {
	if !strings.Contains(selector, "*") {
		return name == selector
	}
	if strings.HasPrefix(selector, "*") {
		return strings.HasSuffix(name, selector[1:])
	}
	return false
}
