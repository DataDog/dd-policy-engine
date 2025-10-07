// Code generated from ./internal/parser/QueryParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // QueryParser
import "github.com/antlr4-go/antlr/v4"

// BaseQueryParserListener is a complete listener for a parse tree produced by QueryParser.
type BaseQueryParserListener struct{}

var _ QueryParserListener = &BaseQueryParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseQueryParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseQueryParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseQueryParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseQueryParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterTopLevelQuery is called when production topLevelQuery is entered.
func (s *BaseQueryParserListener) EnterTopLevelQuery(ctx *TopLevelQueryContext) {}

// ExitTopLevelQuery is called when production topLevelQuery is exited.
func (s *BaseQueryParserListener) ExitTopLevelQuery(ctx *TopLevelQueryContext) {}

// EnterOrExpr is called when production orExpr is entered.
func (s *BaseQueryParserListener) EnterOrExpr(ctx *OrExprContext) {}

// ExitOrExpr is called when production orExpr is exited.
func (s *BaseQueryParserListener) ExitOrExpr(ctx *OrExprContext) {}

// EnterAndExpr is called when production andExpr is entered.
func (s *BaseQueryParserListener) EnterAndExpr(ctx *AndExprContext) {}

// ExitAndExpr is called when production andExpr is exited.
func (s *BaseQueryParserListener) ExitAndExpr(ctx *AndExprContext) {}

// EnterModClause is called when production modClause is entered.
func (s *BaseQueryParserListener) EnterModClause(ctx *ModClauseContext) {}

// ExitModClause is called when production modClause is exited.
func (s *BaseQueryParserListener) ExitModClause(ctx *ModClauseContext) {}

// EnterModifiers is called when production modifiers is entered.
func (s *BaseQueryParserListener) EnterModifiers(ctx *ModifiersContext) {}

// ExitModifiers is called when production modifiers is exited.
func (s *BaseQueryParserListener) ExitModifiers(ctx *ModifiersContext) {}

// EnterClause is called when production clause is entered.
func (s *BaseQueryParserListener) EnterClause(ctx *ClauseContext) {}

// ExitClause is called when production clause is exited.
func (s *BaseQueryParserListener) ExitClause(ctx *ClauseContext) {}

// EnterField is called when production field is entered.
func (s *BaseQueryParserListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BaseQueryParserListener) ExitField(ctx *FieldContext) {}

// EnterSimpleValue is called when production simpleValue is entered.
func (s *BaseQueryParserListener) EnterSimpleValue(ctx *SimpleValueContext) {}

// ExitSimpleValue is called when production simpleValue is exited.
func (s *BaseQueryParserListener) ExitSimpleValue(ctx *SimpleValueContext) {}

// EnterNormal is called when production normal is entered.
func (s *BaseQueryParserListener) EnterNormal(ctx *NormalContext) {}

// ExitNormal is called when production normal is exited.
func (s *BaseQueryParserListener) ExitNormal(ctx *NormalContext) {}

// EnterPrefix is called when production prefix is entered.
func (s *BaseQueryParserListener) EnterPrefix(ctx *PrefixContext) {}

// ExitPrefix is called when production prefix is exited.
func (s *BaseQueryParserListener) ExitPrefix(ctx *PrefixContext) {}

// EnterSuffix is called when production suffix is entered.
func (s *BaseQueryParserListener) EnterSuffix(ctx *SuffixContext) {}

// ExitSuffix is called when production suffix is exited.
func (s *BaseQueryParserListener) ExitSuffix(ctx *SuffixContext) {}

// EnterContains is called when production contains is entered.
func (s *BaseQueryParserListener) EnterContains(ctx *ContainsContext) {}

// ExitContains is called when production contains is exited.
func (s *BaseQueryParserListener) ExitContains(ctx *ContainsContext) {}
