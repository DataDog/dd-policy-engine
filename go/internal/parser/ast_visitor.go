package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// ASTVisitor implements the QueryParserVisitor interface to build AST nodes
// from ANTLR parse trees.
type ASTVisitor struct {
	*BaseQueryParserVisitor
}

// NewASTVisitor creates a new AST visitor instance.
func NewASTVisitor() *ASTVisitor {
	return &ASTVisitor{}
}

// VisitTopLevelQuery visits the top-level query rule and returns the root AST node.
func (v *ASTVisitor) VisitTopLevelQuery(ctx *TopLevelQueryContext) interface{} {
	if ctx.OrExpr() != nil {
		return ctx.OrExpr().Accept(v)
	}
	return &EmptyNode{}
}

// VisitQuery visits a query rule and builds the appropriate AST node.
func (v *ASTVisitor) VisitOrExpr(ctx *OrExprContext) interface{} {
	andExprs := ctx.AllAndExpr()
	if len(andExprs) == 0 {
		return &EmptyNode{}
	}

	// Build a chain of or operations
	currentNode := andExprs[0].Accept(v)
	for _, andExpr := range andExprs[1:] {
		rightNode := andExpr.Accept(v)
		currentNode = &BooleanNode{
			Left:  currentNode,
			Right: rightNode,
			Kind:  Or,
		}
	}
	return currentNode
}

// VisitFirstClause visits a first clause rule.
func (v *ASTVisitor) VisitAndExpr(ctx *AndExprContext) interface{} {
	modClauses := ctx.AllModClause()
	if len(modClauses) == 0 {
		return &EmptyNode{}
	}

	// Build a chain of or operations
	currentNode := modClauses[0].Accept(v)
	for _, modClause := range modClauses[1:] {
		rightNode := modClause.Accept(v)
		currentNode = &BooleanNode{
			Left:  currentNode,
			Right: rightNode,
			Kind:  And,
		}
	}
	return currentNode
}

// VisitConjunctionClause visits a conjunction clause rule.
func (v *ASTVisitor) VisitModClause(ctx *ModClauseContext) interface{} {
	node := ctx.Clause().Accept(v).(Node)

	// Apply NOT modifier if present
	modifiers := ctx.Modifiers()
	if modifiers != nil && modifiers.NOT() != nil {
		node = &UnaryBooleanNode{
			Expr: node,
			Kind: Not,
		}
	}
	return node
}

// VisitModifiers visits a modifiers rule.
func (v *ASTVisitor) VisitModifiers(ctx *ModifiersContext) interface{} {
	// This is handled by the parent clause visitors
	return EmptyNode{}
}

// VisitClause visits a clause rule and builds the appropriate AST node.
func (v *ASTVisitor) VisitClause(ctx *ClauseContext) interface{} {
	// Check for parenthesized query
	if ctx.LPAREN() != nil && ctx.RPAREN() != nil && ctx.OrExpr() != nil {
		return ctx.OrExpr().Accept(v)
	}

	// Check for field-value pair
	if ctx.Field() == nil || ctx.SimpleValue() == nil {
		return &EmptyNode{}
	}

	field := ctx.Field().Accept(v).(string)
	value := ctx.SimpleValue().Accept(v).(ValueResult)
	return &TermNode{
		Field:      field,
		Value:      value.Value,
		Comparator: value.Comparator,
	}
}

// VisitField visits a field rule and returns the field name.
func (v *ASTVisitor) VisitField(ctx *FieldContext) interface{} {
	if ctx.TERM() == nil {
		return ""
	}
	return ctx.TERM().GetText()
}

// VisitSimpleValue visits a simple value rule.
func (v *ASTVisitor) VisitSimpleValue(ctx *SimpleValueContext) interface{} {
	if ctx.Normal() != nil {
		return ctx.Normal().Accept(v)
	}
	if ctx.Prefix() != nil {
		return ctx.Prefix().Accept(v)
	}
	if ctx.Suffix() != nil {
		return ctx.Suffix().Accept(v)
	}
	if ctx.Contains() != nil {
		return ctx.Contains().Accept(v)
	}
	return EmptyNode{}
}

// VisitNormal visits a normal term rule.
func (v *ASTVisitor) VisitNormal(ctx *NormalContext) interface{} {
	if ctx.TERM() == nil {
		return ValueResult{Value: "", Comparator: Exact}
	}
	return ValueResult{Value: ctx.TERM().GetText(), Comparator: Exact}
}

// VisitPrefix visits a prefix term rule.
func (v *ASTVisitor) VisitPrefix(ctx *PrefixContext) interface{} {
	if ctx.TERM_PREFIX() == nil {
		return ValueResult{Value: "", Comparator: Prefix}
	}
	text := ctx.TERM_PREFIX().GetText()
	// Remove the trailing * for the value
	text = text[:len(text)-1]
	return ValueResult{Value: text, Comparator: Prefix}
}

// VisitSuffix visits a suffix term rule.
func (v *ASTVisitor) VisitSuffix(ctx *SuffixContext) interface{} {
	if ctx.TERM_SUFFIX() == nil {
		return ValueResult{Value: "", Comparator: Suffix}
	}
	text := ctx.TERM_SUFFIX().GetText()
	// Remove the leading * for the value
	text = text[1:]
	return ValueResult{Value: text, Comparator: Suffix}
}

// VisitContains visits a contains term rule.
func (v *ASTVisitor) VisitContains(ctx *ContainsContext) interface{} {
	if ctx.TERM_CONTAINS() == nil {
		return ValueResult{Value: "", Comparator: Contains}
	}
	text := ctx.TERM_CONTAINS().GetText()
	// Remove the leading and trailing * for the value
	text = text[1 : len(text)-1]
	return ValueResult{Value: text, Comparator: Contains}
}

// ValueResult represents the result of parsing a value with its comparator and operator.
type ValueResult struct {
	Value      string
	Comparator Comparator
}

// Parse parses an expression.
func Parse(query string) (AST, error) {
	// Create input stream
	input := antlr.NewInputStream(query)

	// Create lexer
	lexer := NewQueryLexer(input)

	// Add error listener to capture lexer errors
	lexer.RemoveErrorListeners()
	errorListener := NewErrorListener()
	lexer.AddErrorListener(errorListener)

	// Create token stream
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create parser
	parser := NewQueryParser(stream)

	// Add error listener to capture parser errors
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	// Parse the input
	tree := parser.TopLevelQuery()

	// Check for errors
	if errorListener.HasErrors() {
		return AST{}, fmt.Errorf("parse error: %s", errorListener.GetErrorString())
	}

	// Create visitor and build AST
	visitor := NewASTVisitor()
	result := tree.Accept(visitor)

	return AST{Node: result}, nil
}

// ErrorListener captures parsing errors.
type ErrorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

// NewErrorListener creates a new error listener.
func NewErrorListener() *ErrorListener {
	return &ErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		errors:               make([]string, 0),
	}
}

// SyntaxError captures syntax errors.
func (el *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	el.errors = append(el.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

// HasErrors returns true if there are any errors.
func (el *ErrorListener) HasErrors() bool {
	return len(el.errors) > 0
}

// GetErrorString returns a formatted error string.
func (el *ErrorListener) GetErrorString() string {
	return strings.Join(el.errors, "; ")
}

var _ QueryParserVisitor = (*ASTVisitor)(nil)
