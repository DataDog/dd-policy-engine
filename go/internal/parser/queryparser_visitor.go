// Code generated from ./internal/parser/QueryParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // QueryParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by QueryParser.
type QueryParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by QueryParser#topLevelQuery.
	VisitTopLevelQuery(ctx *TopLevelQueryContext) interface{}

	// Visit a parse tree produced by QueryParser#orExpr.
	VisitOrExpr(ctx *OrExprContext) interface{}

	// Visit a parse tree produced by QueryParser#andExpr.
	VisitAndExpr(ctx *AndExprContext) interface{}

	// Visit a parse tree produced by QueryParser#modClause.
	VisitModClause(ctx *ModClauseContext) interface{}

	// Visit a parse tree produced by QueryParser#modifiers.
	VisitModifiers(ctx *ModifiersContext) interface{}

	// Visit a parse tree produced by QueryParser#clause.
	VisitClause(ctx *ClauseContext) interface{}

	// Visit a parse tree produced by QueryParser#field.
	VisitField(ctx *FieldContext) interface{}

	// Visit a parse tree produced by QueryParser#simpleValue.
	VisitSimpleValue(ctx *SimpleValueContext) interface{}

	// Visit a parse tree produced by QueryParser#normal.
	VisitNormal(ctx *NormalContext) interface{}

	// Visit a parse tree produced by QueryParser#prefix.
	VisitPrefix(ctx *PrefixContext) interface{}

	// Visit a parse tree produced by QueryParser#suffix.
	VisitSuffix(ctx *SuffixContext) interface{}

	// Visit a parse tree produced by QueryParser#contains.
	VisitContains(ctx *ContainsContext) interface{}
}
