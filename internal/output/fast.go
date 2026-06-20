package output

import (
	"fmt"
	"io"

	"github.com/nolight132/nls/internal/listing"
)

// RenderNames writes one name per line using the fast listing path.
func RenderNames(w io.Writer, names []string) error {
	for _, name := range names {
		if _, err := fmt.Fprintln(w, name); err != nil {
			return err
		}
	}
	return nil
}

// RenderFast uses the minimal path for native ls-compatible piped output.
func RenderFast(w io.Writer, paths []string, opts listing.Options, outOpts Options) error {
	if listing.CanFastList(opts) && len(paths) == 1 && outOpts.Plain == PlainOne && !outOpts.JSON {
		names, err := listing.FastListNames(paths[0], opts)
		if err != nil {
			return err
		}
		return RenderNames(w, names)
	}

	blocks, err := listing.List(paths, opts)
	if err != nil {
		return err
	}
	return Render(w, blocks, outOpts)
}
