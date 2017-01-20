package print

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

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

	plainXact, err := template.New("plain_xact").Funcs(funcsPlainXact(p.MinimumAccountWidth, p.PostingsIndent)).Parse(tplPlainXact)
	if err != nil {
		return err
	}

	for _, nodeIface := range tree.Root.Nodes {
		switch node := nodeIface.(type) {
		case *parse.XactNode:

			if err = plainXact.Execute(buf, node); err != nil {
				return err
			}
		case *parse.CommentNode:
			_, err := buf.WriteString(node.Comment + "\n")
			if err != nil {
				return err
			}

		case *parse.SpaceNode:
			_, err := buf.WriteString(node.Space)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unprintable node type %T", nodeIface)
		}
	}
	return nil
}
