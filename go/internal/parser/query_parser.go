// Code generated from ./internal/parser/QueryParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // QueryParser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type QueryParser struct {
	*antlr.BaseParser
}

var QueryParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func queryparserParserInit() {
	staticData := &QueryParserParserStaticData
	staticData.LiteralNames = []string{
		"", "'('", "')'", "':'", "'*'", "'\"'", "", "", "", "'<'", "'>'", "'<='",
		"'>='",
	}
	staticData.SymbolicNames = []string{
		"", "LPAREN", "RPAREN", "COLON", "STAR", "DQUOTE", "AND", "OR", "NOT",
		"LT", "GT", "LT_EQ", "GT_EQ", "TERM", "TERM_PREFIX", "TERM_SUFFIX",
		"TERM_CONTAINS", "PHRASE", "WS",
	}
	staticData.RuleNames = []string{
		"topLevelQuery", "orExpr", "andExpr", "modClause", "modifiers", "clause",
		"field", "simpleValue", "normal", "prefix", "suffix", "contains",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 18, 86, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 5, 1, 31, 8, 1, 10,
		1, 12, 1, 34, 9, 1, 1, 2, 1, 2, 1, 2, 5, 2, 39, 8, 2, 10, 2, 12, 2, 42,
		9, 2, 1, 2, 1, 2, 5, 2, 46, 8, 2, 10, 2, 12, 2, 49, 9, 2, 3, 2, 51, 8,
		2, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 3, 4, 58, 8, 4, 1, 5, 1, 5, 1, 5, 1, 5,
		1, 5, 1, 5, 1, 5, 3, 5, 67, 8, 5, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1,
		7, 3, 7, 76, 8, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1, 11,
		1, 11, 0, 0, 12, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 0, 0, 82, 0,
		24, 1, 0, 0, 0, 2, 27, 1, 0, 0, 0, 4, 50, 1, 0, 0, 0, 6, 52, 1, 0, 0, 0,
		8, 57, 1, 0, 0, 0, 10, 66, 1, 0, 0, 0, 12, 68, 1, 0, 0, 0, 14, 75, 1, 0,
		0, 0, 16, 77, 1, 0, 0, 0, 18, 79, 1, 0, 0, 0, 20, 81, 1, 0, 0, 0, 22, 83,
		1, 0, 0, 0, 24, 25, 3, 2, 1, 0, 25, 26, 5, 0, 0, 1, 26, 1, 1, 0, 0, 0,
		27, 32, 3, 4, 2, 0, 28, 29, 5, 7, 0, 0, 29, 31, 3, 4, 2, 0, 30, 28, 1,
		0, 0, 0, 31, 34, 1, 0, 0, 0, 32, 30, 1, 0, 0, 0, 32, 33, 1, 0, 0, 0, 33,
		3, 1, 0, 0, 0, 34, 32, 1, 0, 0, 0, 35, 40, 3, 6, 3, 0, 36, 37, 5, 6, 0,
		0, 37, 39, 3, 6, 3, 0, 38, 36, 1, 0, 0, 0, 39, 42, 1, 0, 0, 0, 40, 38,
		1, 0, 0, 0, 40, 41, 1, 0, 0, 0, 41, 51, 1, 0, 0, 0, 42, 40, 1, 0, 0, 0,
		43, 47, 3, 6, 3, 0, 44, 46, 3, 6, 3, 0, 45, 44, 1, 0, 0, 0, 46, 49, 1,
		0, 0, 0, 47, 45, 1, 0, 0, 0, 47, 48, 1, 0, 0, 0, 48, 51, 1, 0, 0, 0, 49,
		47, 1, 0, 0, 0, 50, 35, 1, 0, 0, 0, 50, 43, 1, 0, 0, 0, 51, 5, 1, 0, 0,
		0, 52, 53, 3, 8, 4, 0, 53, 54, 3, 10, 5, 0, 54, 7, 1, 0, 0, 0, 55, 58,
		5, 8, 0, 0, 56, 58, 1, 0, 0, 0, 57, 55, 1, 0, 0, 0, 57, 56, 1, 0, 0, 0,
		58, 9, 1, 0, 0, 0, 59, 60, 3, 12, 6, 0, 60, 61, 3, 14, 7, 0, 61, 67, 1,
		0, 0, 0, 62, 63, 5, 1, 0, 0, 63, 64, 3, 2, 1, 0, 64, 65, 5, 2, 0, 0, 65,
		67, 1, 0, 0, 0, 66, 59, 1, 0, 0, 0, 66, 62, 1, 0, 0, 0, 67, 11, 1, 0, 0,
		0, 68, 69, 5, 13, 0, 0, 69, 70, 5, 3, 0, 0, 70, 13, 1, 0, 0, 0, 71, 76,
		3, 16, 8, 0, 72, 76, 3, 18, 9, 0, 73, 76, 3, 20, 10, 0, 74, 76, 3, 22,
		11, 0, 75, 71, 1, 0, 0, 0, 75, 72, 1, 0, 0, 0, 75, 73, 1, 0, 0, 0, 75,
		74, 1, 0, 0, 0, 76, 15, 1, 0, 0, 0, 77, 78, 5, 13, 0, 0, 78, 17, 1, 0,
		0, 0, 79, 80, 5, 14, 0, 0, 80, 19, 1, 0, 0, 0, 81, 82, 5, 15, 0, 0, 82,
		21, 1, 0, 0, 0, 83, 84, 5, 16, 0, 0, 84, 23, 1, 0, 0, 0, 7, 32, 40, 47,
		50, 57, 66, 75,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// QueryParserInit initializes any static state used to implement QueryParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewQueryParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func QueryParserInit() {
	staticData := &QueryParserParserStaticData
	staticData.once.Do(queryparserParserInit)
}

// NewQueryParser produces a new parser instance for the optional input antlr.TokenStream.
func NewQueryParser(input antlr.TokenStream) *QueryParser {
	QueryParserInit()
	this := new(QueryParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &QueryParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "QueryParser.g4"

	return this
}

// QueryParser tokens.
const (
	QueryParserEOF           = antlr.TokenEOF
	QueryParserLPAREN        = 1
	QueryParserRPAREN        = 2
	QueryParserCOLON         = 3
	QueryParserSTAR          = 4
	QueryParserDQUOTE        = 5
	QueryParserAND           = 6
	QueryParserOR            = 7
	QueryParserNOT           = 8
	QueryParserLT            = 9
	QueryParserGT            = 10
	QueryParserLT_EQ         = 11
	QueryParserGT_EQ         = 12
	QueryParserTERM          = 13
	QueryParserTERM_PREFIX   = 14
	QueryParserTERM_SUFFIX   = 15
	QueryParserTERM_CONTAINS = 16
	QueryParserPHRASE        = 17
	QueryParserWS            = 18
)

// QueryParser rules.
const (
	QueryParserRULE_topLevelQuery = 0
	QueryParserRULE_orExpr        = 1
	QueryParserRULE_andExpr       = 2
	QueryParserRULE_modClause     = 3
	QueryParserRULE_modifiers     = 4
	QueryParserRULE_clause        = 5
	QueryParserRULE_field         = 6
	QueryParserRULE_simpleValue   = 7
	QueryParserRULE_normal        = 8
	QueryParserRULE_prefix        = 9
	QueryParserRULE_suffix        = 10
	QueryParserRULE_contains      = 11
)

// ITopLevelQueryContext is an interface to support dynamic dispatch.
type ITopLevelQueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	OrExpr() IOrExprContext
	EOF() antlr.TerminalNode

	// IsTopLevelQueryContext differentiates from other interfaces.
	IsTopLevelQueryContext()
}

type TopLevelQueryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTopLevelQueryContext() *TopLevelQueryContext {
	var p = new(TopLevelQueryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_topLevelQuery
	return p
}

func InitEmptyTopLevelQueryContext(p *TopLevelQueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_topLevelQuery
}

func (*TopLevelQueryContext) IsTopLevelQueryContext() {}

func NewTopLevelQueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TopLevelQueryContext {
	var p = new(TopLevelQueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_topLevelQuery

	return p
}

func (s *TopLevelQueryContext) GetParser() antlr.Parser { return s.parser }

func (s *TopLevelQueryContext) OrExpr() IOrExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrExprContext)
}

func (s *TopLevelQueryContext) EOF() antlr.TerminalNode {
	return s.GetToken(QueryParserEOF, 0)
}

func (s *TopLevelQueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TopLevelQueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TopLevelQueryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterTopLevelQuery(s)
	}
}

func (s *TopLevelQueryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitTopLevelQuery(s)
	}
}

func (s *TopLevelQueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitTopLevelQuery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) TopLevelQuery() (localctx ITopLevelQueryContext) {
	localctx = NewTopLevelQueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, QueryParserRULE_topLevelQuery)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(24)
		p.OrExpr()
	}
	{
		p.SetState(25)
		p.Match(QueryParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrExprContext is an interface to support dynamic dispatch.
type IOrExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAndExpr() []IAndExprContext
	AndExpr(i int) IAndExprContext
	AllOR() []antlr.TerminalNode
	OR(i int) antlr.TerminalNode

	// IsOrExprContext differentiates from other interfaces.
	IsOrExprContext()
}

type OrExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrExprContext() *OrExprContext {
	var p = new(OrExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_orExpr
	return p
}

func InitEmptyOrExprContext(p *OrExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_orExpr
}

func (*OrExprContext) IsOrExprContext() {}

func NewOrExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrExprContext {
	var p = new(OrExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_orExpr

	return p
}

func (s *OrExprContext) GetParser() antlr.Parser { return s.parser }

func (s *OrExprContext) AllAndExpr() []IAndExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAndExprContext); ok {
			len++
		}
	}

	tst := make([]IAndExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAndExprContext); ok {
			tst[i] = t.(IAndExprContext)
			i++
		}
	}

	return tst
}

func (s *OrExprContext) AndExpr(i int) IAndExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAndExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAndExprContext)
}

func (s *OrExprContext) AllOR() []antlr.TerminalNode {
	return s.GetTokens(QueryParserOR)
}

func (s *OrExprContext) OR(i int) antlr.TerminalNode {
	return s.GetToken(QueryParserOR, i)
}

func (s *OrExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterOrExpr(s)
	}
}

func (s *OrExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitOrExpr(s)
	}
}

func (s *OrExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitOrExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) OrExpr() (localctx IOrExprContext) {
	localctx = NewOrExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, QueryParserRULE_orExpr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(27)
		p.AndExpr()
	}
	p.SetState(32)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == QueryParserOR {
		{
			p.SetState(28)
			p.Match(QueryParserOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(29)
			p.AndExpr()
		}

		p.SetState(34)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAndExprContext is an interface to support dynamic dispatch.
type IAndExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllModClause() []IModClauseContext
	ModClause(i int) IModClauseContext
	AllAND() []antlr.TerminalNode
	AND(i int) antlr.TerminalNode

	// IsAndExprContext differentiates from other interfaces.
	IsAndExprContext()
}

type AndExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAndExprContext() *AndExprContext {
	var p = new(AndExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_andExpr
	return p
}

func InitEmptyAndExprContext(p *AndExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_andExpr
}

func (*AndExprContext) IsAndExprContext() {}

func NewAndExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AndExprContext {
	var p = new(AndExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_andExpr

	return p
}

func (s *AndExprContext) GetParser() antlr.Parser { return s.parser }

func (s *AndExprContext) AllModClause() []IModClauseContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IModClauseContext); ok {
			len++
		}
	}

	tst := make([]IModClauseContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IModClauseContext); ok {
			tst[i] = t.(IModClauseContext)
			i++
		}
	}

	return tst
}

func (s *AndExprContext) ModClause(i int) IModClauseContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IModClauseContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IModClauseContext)
}

func (s *AndExprContext) AllAND() []antlr.TerminalNode {
	return s.GetTokens(QueryParserAND)
}

func (s *AndExprContext) AND(i int) antlr.TerminalNode {
	return s.GetToken(QueryParserAND, i)
}

func (s *AndExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AndExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterAndExpr(s)
	}
}

func (s *AndExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitAndExpr(s)
	}
}

func (s *AndExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitAndExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) AndExpr() (localctx IAndExprContext) {
	localctx = NewAndExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, QueryParserRULE_andExpr)
	var _la int

	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(35)
			p.ModClause()
		}
		p.SetState(40)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == QueryParserAND {
			{
				p.SetState(36)
				p.Match(QueryParserAND)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(37)
				p.ModClause()
			}

			p.SetState(42)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(43)
			p.ModClause()
		}
		p.SetState(47)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&8450) != 0 {
			{
				p.SetState(44)
				p.ModClause()
			}

			p.SetState(49)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IModClauseContext is an interface to support dynamic dispatch.
type IModClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Modifiers() IModifiersContext
	Clause() IClauseContext

	// IsModClauseContext differentiates from other interfaces.
	IsModClauseContext()
}

type ModClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyModClauseContext() *ModClauseContext {
	var p = new(ModClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_modClause
	return p
}

func InitEmptyModClauseContext(p *ModClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_modClause
}

func (*ModClauseContext) IsModClauseContext() {}

func NewModClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ModClauseContext {
	var p = new(ModClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_modClause

	return p
}

func (s *ModClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ModClauseContext) Modifiers() IModifiersContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IModifiersContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IModifiersContext)
}

func (s *ModClauseContext) Clause() IClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IClauseContext)
}

func (s *ModClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ModClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ModClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterModClause(s)
	}
}

func (s *ModClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitModClause(s)
	}
}

func (s *ModClauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitModClause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) ModClause() (localctx IModClauseContext) {
	localctx = NewModClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, QueryParserRULE_modClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(52)
		p.Modifiers()
	}
	{
		p.SetState(53)
		p.Clause()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IModifiersContext is an interface to support dynamic dispatch.
type IModifiersContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NOT() antlr.TerminalNode

	// IsModifiersContext differentiates from other interfaces.
	IsModifiersContext()
}

type ModifiersContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyModifiersContext() *ModifiersContext {
	var p = new(ModifiersContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_modifiers
	return p
}

func InitEmptyModifiersContext(p *ModifiersContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_modifiers
}

func (*ModifiersContext) IsModifiersContext() {}

func NewModifiersContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ModifiersContext {
	var p = new(ModifiersContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_modifiers

	return p
}

func (s *ModifiersContext) GetParser() antlr.Parser { return s.parser }

func (s *ModifiersContext) NOT() antlr.TerminalNode {
	return s.GetToken(QueryParserNOT, 0)
}

func (s *ModifiersContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ModifiersContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ModifiersContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterModifiers(s)
	}
}

func (s *ModifiersContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitModifiers(s)
	}
}

func (s *ModifiersContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitModifiers(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Modifiers() (localctx IModifiersContext) {
	localctx = NewModifiersContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, QueryParserRULE_modifiers)
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryParserNOT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(55)
			p.Match(QueryParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case QueryParserLPAREN, QueryParserTERM:
		p.EnterOuterAlt(localctx, 2)

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IClauseContext is an interface to support dynamic dispatch.
type IClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Field() IFieldContext
	SimpleValue() ISimpleValueContext
	LPAREN() antlr.TerminalNode
	OrExpr() IOrExprContext
	RPAREN() antlr.TerminalNode

	// IsClauseContext differentiates from other interfaces.
	IsClauseContext()
}

type ClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyClauseContext() *ClauseContext {
	var p = new(ClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_clause
	return p
}

func InitEmptyClauseContext(p *ClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_clause
}

func (*ClauseContext) IsClauseContext() {}

func NewClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ClauseContext {
	var p = new(ClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_clause

	return p
}

func (s *ClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ClauseContext) Field() IFieldContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldContext)
}

func (s *ClauseContext) SimpleValue() ISimpleValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISimpleValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISimpleValueContext)
}

func (s *ClauseContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(QueryParserLPAREN, 0)
}

func (s *ClauseContext) OrExpr() IOrExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrExprContext)
}

func (s *ClauseContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(QueryParserRPAREN, 0)
}

func (s *ClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterClause(s)
	}
}

func (s *ClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitClause(s)
	}
}

func (s *ClauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitClause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Clause() (localctx IClauseContext) {
	localctx = NewClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, QueryParserRULE_clause)
	p.SetState(66)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryParserTERM:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(59)
			p.Field()
		}

		{
			p.SetState(60)
			p.SimpleValue()
		}

	case QueryParserLPAREN:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(62)
			p.Match(QueryParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(63)
			p.OrExpr()
		}
		{
			p.SetState(64)
			p.Match(QueryParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldContext is an interface to support dynamic dispatch.
type IFieldContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM() antlr.TerminalNode
	COLON() antlr.TerminalNode

	// IsFieldContext differentiates from other interfaces.
	IsFieldContext()
}

type FieldContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldContext() *FieldContext {
	var p = new(FieldContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_field
	return p
}

func InitEmptyFieldContext(p *FieldContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_field
}

func (*FieldContext) IsFieldContext() {}

func NewFieldContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldContext {
	var p = new(FieldContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_field

	return p
}

func (s *FieldContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldContext) TERM() antlr.TerminalNode {
	return s.GetToken(QueryParserTERM, 0)
}

func (s *FieldContext) COLON() antlr.TerminalNode {
	return s.GetToken(QueryParserCOLON, 0)
}

func (s *FieldContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterField(s)
	}
}

func (s *FieldContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitField(s)
	}
}

func (s *FieldContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitField(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Field() (localctx IFieldContext) {
	localctx = NewFieldContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, QueryParserRULE_field)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(68)
		p.Match(QueryParserTERM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(69)
		p.Match(QueryParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISimpleValueContext is an interface to support dynamic dispatch.
type ISimpleValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Normal() INormalContext
	Prefix() IPrefixContext
	Suffix() ISuffixContext
	Contains() IContainsContext

	// IsSimpleValueContext differentiates from other interfaces.
	IsSimpleValueContext()
}

type SimpleValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimpleValueContext() *SimpleValueContext {
	var p = new(SimpleValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_simpleValue
	return p
}

func InitEmptySimpleValueContext(p *SimpleValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_simpleValue
}

func (*SimpleValueContext) IsSimpleValueContext() {}

func NewSimpleValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SimpleValueContext {
	var p = new(SimpleValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_simpleValue

	return p
}

func (s *SimpleValueContext) GetParser() antlr.Parser { return s.parser }

func (s *SimpleValueContext) Normal() INormalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INormalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INormalContext)
}

func (s *SimpleValueContext) Prefix() IPrefixContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrefixContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrefixContext)
}

func (s *SimpleValueContext) Suffix() ISuffixContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISuffixContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISuffixContext)
}

func (s *SimpleValueContext) Contains() IContainsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IContainsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IContainsContext)
}

func (s *SimpleValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SimpleValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SimpleValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterSimpleValue(s)
	}
}

func (s *SimpleValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitSimpleValue(s)
	}
}

func (s *SimpleValueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitSimpleValue(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) SimpleValue() (localctx ISimpleValueContext) {
	localctx = NewSimpleValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, QueryParserRULE_simpleValue)
	p.SetState(75)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case QueryParserTERM:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(71)
			p.Normal()
		}

	case QueryParserTERM_PREFIX:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(72)
			p.Prefix()
		}

	case QueryParserTERM_SUFFIX:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(73)
			p.Suffix()
		}

	case QueryParserTERM_CONTAINS:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(74)
			p.Contains()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INormalContext is an interface to support dynamic dispatch.
type INormalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM() antlr.TerminalNode

	// IsNormalContext differentiates from other interfaces.
	IsNormalContext()
}

type NormalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNormalContext() *NormalContext {
	var p = new(NormalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_normal
	return p
}

func InitEmptyNormalContext(p *NormalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_normal
}

func (*NormalContext) IsNormalContext() {}

func NewNormalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NormalContext {
	var p = new(NormalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_normal

	return p
}

func (s *NormalContext) GetParser() antlr.Parser { return s.parser }

func (s *NormalContext) TERM() antlr.TerminalNode {
	return s.GetToken(QueryParserTERM, 0)
}

func (s *NormalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NormalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NormalContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterNormal(s)
	}
}

func (s *NormalContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitNormal(s)
	}
}

func (s *NormalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitNormal(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Normal() (localctx INormalContext) {
	localctx = NewNormalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, QueryParserRULE_normal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(77)
		p.Match(QueryParserTERM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPrefixContext is an interface to support dynamic dispatch.
type IPrefixContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_PREFIX() antlr.TerminalNode

	// IsPrefixContext differentiates from other interfaces.
	IsPrefixContext()
}

type PrefixContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPrefixContext() *PrefixContext {
	var p = new(PrefixContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_prefix
	return p
}

func InitEmptyPrefixContext(p *PrefixContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_prefix
}

func (*PrefixContext) IsPrefixContext() {}

func NewPrefixContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PrefixContext {
	var p = new(PrefixContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_prefix

	return p
}

func (s *PrefixContext) GetParser() antlr.Parser { return s.parser }

func (s *PrefixContext) TERM_PREFIX() antlr.TerminalNode {
	return s.GetToken(QueryParserTERM_PREFIX, 0)
}

func (s *PrefixContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrefixContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PrefixContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterPrefix(s)
	}
}

func (s *PrefixContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitPrefix(s)
	}
}

func (s *PrefixContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitPrefix(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Prefix() (localctx IPrefixContext) {
	localctx = NewPrefixContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, QueryParserRULE_prefix)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(79)
		p.Match(QueryParserTERM_PREFIX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISuffixContext is an interface to support dynamic dispatch.
type ISuffixContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_SUFFIX() antlr.TerminalNode

	// IsSuffixContext differentiates from other interfaces.
	IsSuffixContext()
}

type SuffixContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySuffixContext() *SuffixContext {
	var p = new(SuffixContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_suffix
	return p
}

func InitEmptySuffixContext(p *SuffixContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_suffix
}

func (*SuffixContext) IsSuffixContext() {}

func NewSuffixContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SuffixContext {
	var p = new(SuffixContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_suffix

	return p
}

func (s *SuffixContext) GetParser() antlr.Parser { return s.parser }

func (s *SuffixContext) TERM_SUFFIX() antlr.TerminalNode {
	return s.GetToken(QueryParserTERM_SUFFIX, 0)
}

func (s *SuffixContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SuffixContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SuffixContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterSuffix(s)
	}
}

func (s *SuffixContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitSuffix(s)
	}
}

func (s *SuffixContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitSuffix(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Suffix() (localctx ISuffixContext) {
	localctx = NewSuffixContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, QueryParserRULE_suffix)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(81)
		p.Match(QueryParserTERM_SUFFIX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IContainsContext is an interface to support dynamic dispatch.
type IContainsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_CONTAINS() antlr.TerminalNode

	// IsContainsContext differentiates from other interfaces.
	IsContainsContext()
}

type ContainsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyContainsContext() *ContainsContext {
	var p = new(ContainsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_contains
	return p
}

func InitEmptyContainsContext(p *ContainsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = QueryParserRULE_contains
}

func (*ContainsContext) IsContainsContext() {}

func NewContainsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ContainsContext {
	var p = new(ContainsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = QueryParserRULE_contains

	return p
}

func (s *ContainsContext) GetParser() antlr.Parser { return s.parser }

func (s *ContainsContext) TERM_CONTAINS() antlr.TerminalNode {
	return s.GetToken(QueryParserTERM_CONTAINS, 0)
}

func (s *ContainsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ContainsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ContainsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.EnterContains(s)
	}
}

func (s *ContainsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(QueryParserListener); ok {
		listenerT.ExitContains(s)
	}
}

func (s *ContainsContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case QueryParserVisitor:
		return t.VisitContains(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *QueryParser) Contains() (localctx IContainsContext) {
	localctx = NewContainsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, QueryParserRULE_contains)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(83)
		p.Match(QueryParserTERM_CONTAINS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
