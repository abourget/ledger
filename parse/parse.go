package parse

// Tree is the representation of a single parsed Ledger file.
type Tree struct {
	Name string // name of the template represented by the tree.
	// ParseName -> TopFileName
	TopFileName string    // name of the top-level ledger file during parsing, for error messages.
	Root        *ListNode // top-level root of the tree.
	text        string    // text parsed to create the template (or its parent)
	// Parsing only; cleared after parse.
	funcs     []map[string]interface{}
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
	vars      []string // variables defined at the moment.
	treeSet   map[string]*Tree
}
