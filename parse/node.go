package parse

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

var textFormat = "%s" // Changed to "%q" in tests for better error messages.

// A Node is an element in the parse tree. The interface is trivial.
// The interface contains an unexported method so that only
// types local to this package can satisfy it.
type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	//Copy() Node
	Position() Pos // byte position of start of node in full original input string
	// tree returns the containing *Tree.
	// It is unexported so all implementations of Node are in this package.
	tree() *Tree
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

func (t NodeType) String() string {
	label, found := nodeLabel[t]
	if !found {
		return fmt.Sprintf("%d", t)
	}
	return label
}

const (
	NodeJournal NodeType = iota
	NodeList
	NodeXact
	NodePosting
	NodeComment
	NodeSpace
	NodeAmount
)

var nodeLabel = map[NodeType]string{
	NodeJournal: "NodeJournal",
	NodeList:    "NodeList",
	NodeXact:    "NodeXact",
	NodePosting: "NodePosting",
	NodeComment: "NodeComment",
	NodeSpace:   "NodeSpace",
	NodeAmount:  "NodeAmount",
}

/** ListNode **/

// ListNode holds a sequence of Entry Nodes
type ListNode struct {
	NodeType
	Pos
	tr *Tree

	Nodes []Node // The top-level element nodes in lexical order.
}

func (t *Tree) newList(pos Pos) *ListNode {
	return &ListNode{tr: t, NodeType: NodeList, Pos: pos}
}

func (l *ListNode) add(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) tree() *Tree {
	return l.tr
}

func (l *ListNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}

// func (l *ListNode) CopyList() *ListNode {
// 	if l == nil {
// 		return l
// 	}
// 	n := l.tr.newList(l.Pos)
// 	for _, elem := range l.Nodes {
// 		n.append(elem.Copy())
// 	}
// 	return n
// }

// func (l *ListNode) Copy() Node {
// 	return l.CopyList()
// }

/** SpaceNode **/

type SpaceNode struct {
	NodeType
	Pos
	tr *Tree

	Space string
}

func (t *Tree) newSpace(p Pos, spaces string) *SpaceNode {
	return &SpaceNode{NodeType: NodeSpace, Pos: p, tr: t, Space: spaces}
}

func (n *SpaceNode) String() string { return n.Space }
func (n *SpaceNode) tree() *Tree    { return n.tr }

/** CommentNode **/

type CommentNode struct {
	NodeType
	Pos
	tr *Tree

	Comment string
}

func (t *Tree) newComment(i item) *CommentNode {
	return &CommentNode{NodeType: NodeComment, Pos: i.pos, tr: t, Comment: i.val}
}

func (n *CommentNode) String() string { return n.Comment }
func (n *CommentNode) tree() *Tree    { return n.tr }

/** XactNode - Ledger Transactions **/

type XactNode struct {
	NodeType
	Pos
	tr *Tree

	Date          time.Time
	EffectiveDate time.Time
	Description   string
	Code          string
	IsPending     bool
	IsCleared     bool
	NotePreSpace  string
	Note          string
	Postings      []*PostingNode
}

func (t *Tree) newXact(pos Pos) *XactNode {
	n := &XactNode{tr: t, NodeType: NodeXact, Pos: pos}
	t.Root.add(n)
	return n
}

func (n *XactNode) String() string {
	msg := []string{n.Date.Format("2006-01-02")}
	if n.IsPending {
		msg = append(msg, "!")
	}
	if n.IsCleared {
		msg = append(msg, "*")
	}
	msg = append(msg, n.Description)
	if n.Note != "" {
		msg = append(msg, n.Note)
	}
	return fmt.Sprintf(textFormat, strings.Join(msg, " "))
}

func (n *XactNode) tree() *Tree { return n.tr }

// func (n *XactNode) Copy() Node {
// 	return &XactNode{tr: n.tr, NodeType: n.NodeType, Pos: n.Pos, Date: n.Date, Description: n.Description, IsPending: n.IsPending, IsCleared: n.IsCleared}
// }

func (n *XactNode) newPosting(pos Pos) *PostingNode {
	p := &PostingNode{tr: n.tr, NodeType: NodePosting, Pos: pos}
	n.Postings = append(n.Postings, p)
	return p
}

/** PostingNode - Postings to transactions **/

type PostingNode struct {
	NodeType
	Pos
	tr *Tree

	AccountPreSpace   string
	Account           string
	AccountPostSpace  string // if non-empty, this must always be two spaces, one space and one tab, one tab, or more whitespace, as per specs.
	Amount            *AmountNode
	BalanceAssertion  *AmountNode
	BalanceAssignment *AmountNode
	Price             *AmountNode
	PriceIsForWhole   bool // false = per unit (@); true = price for the whole (@@), meaningful only if `Price` is defined.
	LotDate           time.Time
	LotPrice          *AmountNode
	NotePreSpace      string
	Note              string
}

func (n *PostingNode) String() string {
	msg := []string{n.AccountPreSpace, n.Account, n.AccountPostSpace, n.Amount.String()}
	// TODO: add the other things in here...
	if n.Note != "" {
		msg = append(msg, n.NotePreSpace, n.Note)
	}
	return fmt.Sprintf(textFormat, strings.Join(msg, ""))
}

func (n *PostingNode) tree() *Tree { return n.tr }

/** PostingNode - Postings to transactions **/

type AmountNode struct {
	NodeType
	Pos
	tr *Tree

	Raw       string // Raw representation of the amount, like "- CAD  100.20", with spacing and all. Will be used for printing unless empty, in which case the string will be reconstructed based on the other values herein.
	Quantity  string
	Negative  bool
	Commodity string
	ValueExpr string // Mutually exclusive with "Quantity". Expression, to be evaluated by some engine..
}

func (t *Tree) newAmount() *AmountNode {
	return &AmountNode{tr: t, NodeType: NodeAmount}
}

func (n *AmountNode) String() string {
	if n == nil {
		return ""
	}
	if n.Raw != "" {
		return n.Raw
	}
	if n.ValueExpr != "" {
		return fmt.Sprintf("(%s)", n.ValueExpr)
	}
	out := n.Quantity
	if n.Negative {
		out = "-" + n.Quantity
	}
	if n.Commodity != "" {
		out += " " + n.Commodity
	}
	return out
}

func (n *AmountNode) tree() *Tree { return n.tr }

func (n *AmountNode) next(t *Tree) item {
	it := t.next()
	if n.Pos == 0 {
		n.Pos = it.pos
	}
	n.Raw += it.val
	return it
}

func (n *AmountNode) space(t *Tree) {
	if it := t.peek(); it.typ == itemSpace {
		n.next(t)
	}
}
