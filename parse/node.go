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

const (
	NodeJournal NodeType = iota
	NodeList
	NodeXact
	NodePosting
	NodeComment
	NodeSpace
	NodeNumber
	NodeSymbol
	NodeDate
)

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

func (t *Tree) newSpace(i item) *SpaceNode {
	return &SpaceNode{NodeType: NodeSpace, Pos: i.pos, tr: t, Space: i.val}
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
	tr    *Tree
	items []item

	Date        time.Time
	Description string
	IsPending   bool
	IsCleared   bool
	Note        string
	Postings    []*PostingNode
}

func (t *Tree) newXact(pos Pos) *XactNode {
	n := &XactNode{tr: t, NodeType: NodeXact, Pos: pos}
	t.Root.add(n)
	return n
}

func (n *XactNode) add(i item) {
	n.items = append(n.items, i)
}

func (n *XactNode) String() string {
	msg := []string{n.Date.Format("2016-01-02")}
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
	tr    *Tree
	items []item

	Account           string
	AmountExpr        string
	Amount            string
	BalanceAssertion  string
	BalanceAssignment string
	Note              string
	Price             string
	PriceWhole        string
	LotDate           string
	LotPrice          string
}

func (n *PostingNode) String() string {
	msg := []string{"    ", n.Account, n.Amount, n.AmountExpr}
	if n.Note != "" {
		msg = append(msg, n.Note)
	}
	return fmt.Sprintf(textFormat, strings.Join(msg, " "))
}

func (n *PostingNode) tree() *Tree { return n.tr }
