// Code generated from ./internal/parser/QueryParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // QueryParser
import "github.com/antlr4-go/antlr/v4"

type BaseQueryParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseQueryParserVisitor) VisitTopLevelQuery(ctx *TopLevelQueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitOrExpr(ctx *OrExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitAndExpr(ctx *AndExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitModClause(ctx *ModClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitModifiers(ctx *ModifiersContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitClause(ctx *ClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitField(ctx *FieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitSimpleValue(ctx *SimpleValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitNormal(ctx *NormalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitPrefix(ctx *PrefixContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitSuffix(ctx *SuffixContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseQueryParserVisitor) VisitContains(ctx *ContainsContext) interface{} {
	return v.VisitChildren(ctx)
}
