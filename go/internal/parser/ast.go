package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AST node definitions
type Node any

type EmptyNode struct{}

type Comparator int

const (
	Exact Comparator = iota
	Prefix
	Suffix
	Contains
)

func (comp Comparator) String() string {
	switch comp {
	case Exact:
		return "EXACT"
	case Prefix:
		return "PREFIX"
	case Suffix:
		return "SUFFIX"
	case Contains:
		return "CONTAINS"
	}
	panic("unexpected comparator")
}

func (c Comparator) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

type TermNode struct {
	Field      string     `json:"Field"`
	Value      string     `json:"Value"`
	Comparator Comparator `json:"Comparator"`
}

type BooleanOperator int

const (
	Or BooleanOperator = iota
	And
)

func (b BooleanOperator) String() string {
	switch b {
	case Or:
		return "OR"
	case And:
		return "AND"
	}
	panic("unexpected boolean operator")
}

type UnaryBooleanOperator int

const (
	Not UnaryBooleanOperator = iota
)

func (u UnaryBooleanOperator) String() string {
	switch u {
	case Not:
		return "NOT"
	}
	panic("unexpected unary boolean operator")
}

type BooleanNode struct {
	Left  Node            `json:"Left"`
	Right Node            `json:"Right"`
	Kind  BooleanOperator `json:"Kind"`
}

type UnaryBooleanNode struct {
	Expr Node
	Kind UnaryBooleanOperator `json:"Kind"`
}

func printAST(node Node, indent string) string {
	var sb strings.Builder

	switch n := node.(type) {
	case *TermNode:
		sb.WriteString(fmt.Sprintf("%s Term(key = \"%s\", value = \"%s\", comparator = \"%s\")\n", indent, n.Field, n.Value, n.Comparator))
	case *BooleanNode:
		leftIndent := indent
		if _, ok := n.Left.(*TermNode); ok {
			leftIndent += "  ├──"
		} else {
			leftIndent += "  │"
		}

		sb.WriteString(fmt.Sprintf("%s Predicate(kind = %s):\n", indent, n.Kind))
		sb.WriteString(printAST(n.Left, leftIndent))
		sb.WriteString(printAST(n.Right, indent+"  └──"))
	case *UnaryBooleanNode:
		if _, ok := n.Expr.(*TermNode); ok {
			indent += "  ├──"
		} else {
			indent += "  │"
		}

		sb.WriteString(fmt.Sprintf("%s Predicate(kind = %s):\n", indent, n.Kind))
		sb.WriteString(printAST(n.Expr, indent))
	default:
		sb.WriteString(fmt.Sprintf("%sUnknown node\n", indent))
	}

	return sb.String()
}

type AST struct {
	Node Node
}

func (ast AST) String() string {
	return printAST(ast.Node, "")
}

func (ast AST) MarshalJSON() ([]byte, error) {
	return json.Marshal(ast.Node)
}

func (b BooleanOperator) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *BooleanNode) MarshalJSON() ([]byte, error) {
	type Alias BooleanNode
	return json.Marshal(map[string]*Alias{
		"BooleanNode": (*Alias)(b),
	})
}
