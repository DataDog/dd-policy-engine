// Code generated from ./internal/parser/QueryParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // QueryParser
import "github.com/antlr4-go/antlr/v4"

// QueryParserListener is a complete listener for a parse tree produced by QueryParser.
type QueryParserListener interface {
	antlr.ParseTreeListener

	// EnterTopLevelQuery is called when entering the topLevelQuery production.
	EnterTopLevelQuery(c *TopLevelQueryContext)

	// EnterOrExpr is called when entering the orExpr production.
	EnterOrExpr(c *OrExprContext)

	// EnterAndExpr is called when entering the andExpr production.
	EnterAndExpr(c *AndExprContext)

	// EnterModClause is called when entering the modClause production.
	EnterModClause(c *ModClauseContext)

	// EnterModifiers is called when entering the modifiers production.
	EnterModifiers(c *ModifiersContext)

	// EnterClause is called when entering the clause production.
	EnterClause(c *ClauseContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterSimpleValue is called when entering the simpleValue production.
	EnterSimpleValue(c *SimpleValueContext)

	// EnterNormal is called when entering the normal production.
	EnterNormal(c *NormalContext)

	// EnterPrefix is called when entering the prefix production.
	EnterPrefix(c *PrefixContext)

	// EnterSuffix is called when entering the suffix production.
	EnterSuffix(c *SuffixContext)

	// EnterContains is called when entering the contains production.
	EnterContains(c *ContainsContext)

	// ExitTopLevelQuery is called when exiting the topLevelQuery production.
	ExitTopLevelQuery(c *TopLevelQueryContext)

	// ExitOrExpr is called when exiting the orExpr production.
	ExitOrExpr(c *OrExprContext)

	// ExitAndExpr is called when exiting the andExpr production.
	ExitAndExpr(c *AndExprContext)

	// ExitModClause is called when exiting the modClause production.
	ExitModClause(c *ModClauseContext)

	// ExitModifiers is called when exiting the modifiers production.
	ExitModifiers(c *ModifiersContext)

	// ExitClause is called when exiting the clause production.
	ExitClause(c *ClauseContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitSimpleValue is called when exiting the simpleValue production.
	ExitSimpleValue(c *SimpleValueContext)

	// ExitNormal is called when exiting the normal production.
	ExitNormal(c *NormalContext)

	// ExitPrefix is called when exiting the prefix production.
	ExitPrefix(c *PrefixContext)

	// ExitSuffix is called when exiting the suffix production.
	ExitSuffix(c *SuffixContext)

	// ExitContains is called when exiting the contains production.
	ExitContains(c *ContainsContext)
}
