package listing

// Block is a directory section for plain or recursive output.
type Block struct {
	// Header is the directory label printed before the section; empty for
	// single-directory listings and bare file operands.
	Header string
	// Dir is the directory the entries live in, even when no header is
	// printed. Empty for file operands, whose names are already paths.
	Dir       string
	Entries   []Entry
	Directory bool
	// GitRepo reports whether a git status lookup succeeded for this
	// block's directory; the git column only renders when it did.
	GitRepo bool
}
