package listing

// Block is a directory section for plain or recursive output.
type Block struct {
	Header  string
	Entries []Entry
}
