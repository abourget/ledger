// Package print provides a low-level writer for ledger files.
package print

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/abourget/ledger/parse"
)

// Printer formats the AST of a Ledger file into a properly formatted
// .ledger file.
type Printer struct {
	tree *parse.Tree

	MinimumAccountWidth int
	PostingsIndent      int
}

func New(tree *parse.Tree) *Printer {
	return &Printer{
		tree:                tree,
		MinimumAccountWidth: 48,
		PostingsIndent:      4,
	}
}

func (p *Printer) Print(buf *bytes.Buffer) error {
	tree := p.tree

	if tree.Root == nil {
		return errors.New("parse tree is empty (Root is nil)")
	}

	for _, nodeIface := range tree.Root.Nodes {
		var err error
		switch node := nodeIface.(type) {
		case *parse.XactNode:
			p.writePlainXact(buf, node)
		case *parse.CommentNode:
			_, err = buf.WriteString(node.Comment + "\n")
		case *parse.SpaceNode:
			_, err = buf.WriteString(node.Space)
		case *parse.DirectiveNode:
			_, err = buf.WriteString(node.Raw)
		default:
			return fmt.Errorf("unprintable node type %T", nodeIface)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
