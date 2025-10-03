// Code generated from ./internal/parser/QueryLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type QueryLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var QueryLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func querylexerLexerInit() {
	staticData := &QueryLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE", "COMPARISON",
	}
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
		"LPAREN", "RPAREN", "COLON", "STAR", "DQUOTE", "AND", "OR", "NOT", "LT",
		"GT", "LT_EQ", "GT_EQ", "ESC_CHAR", "STD_CHAR", "TERM", "TERM_START_CHAR",
		"TERM_CHAR", "TERM_PREFIX", "TERM_SUFFIX", "TERM_CONTAINS", "PHRASE",
		"WS", "NON_WS", "NUMERIC_TERM", "DIGIT", "NUMBER", "COMPARISON_TERM",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 18, 196, 6, -1, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3,
		7, 3, 2, 4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9,
		7, 9, 2, 10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7,
		14, 2, 15, 7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19,
		2, 20, 7, 20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2,
		25, 7, 25, 2, 26, 7, 26, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3,
		1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 3, 5, 72, 8, 5, 1, 6, 1, 6, 1,
		6, 1, 6, 3, 6, 78, 8, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 3, 7, 85, 8, 7,
		1, 8, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 1, 10, 1,
		10, 1, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 13,
		1, 13, 1, 14, 1, 14, 5, 14, 112, 8, 14, 10, 14, 12, 14, 115, 9, 14, 1,
		15, 1, 15, 3, 15, 119, 8, 15, 1, 16, 1, 16, 3, 16, 123, 8, 16, 1, 17, 1,
		17, 5, 17, 127, 8, 17, 10, 17, 12, 17, 130, 9, 17, 1, 17, 1, 17, 1, 18,
		1, 18, 4, 18, 136, 8, 18, 11, 18, 12, 18, 137, 1, 19, 1, 19, 4, 19, 142,
		8, 19, 11, 19, 12, 19, 143, 1, 19, 1, 19, 1, 20, 1, 20, 1, 20, 5, 20, 151,
		8, 20, 10, 20, 12, 20, 154, 9, 20, 1, 20, 1, 20, 1, 21, 4, 21, 159, 8,
		21, 11, 21, 12, 21, 160, 1, 21, 1, 21, 1, 22, 1, 22, 1, 23, 1, 23, 1, 23,
		3, 23, 170, 8, 23, 1, 24, 1, 24, 1, 25, 3, 25, 175, 8, 25, 1, 25, 4, 25,
		178, 8, 25, 11, 25, 12, 25, 179, 1, 25, 1, 25, 4, 25, 184, 8, 25, 11, 25,
		12, 25, 185, 3, 25, 188, 8, 25, 1, 26, 1, 26, 3, 26, 192, 8, 26, 1, 26,
		1, 26, 1, 26, 1, 152, 0, 27, 2, 1, 4, 2, 6, 3, 8, 4, 10, 5, 12, 6, 14,
		7, 16, 8, 18, 9, 20, 10, 22, 11, 24, 12, 26, 0, 28, 0, 30, 13, 32, 0, 34,
		0, 36, 14, 38, 15, 40, 16, 42, 17, 44, 18, 46, 0, 48, 0, 50, 0, 52, 0,
		54, 0, 2, 0, 1, 6, 2, 0, 33, 33, 45, 45, 11, 0, 9, 10, 13, 13, 32, 34,
		40, 42, 45, 45, 58, 58, 60, 60, 62, 63, 91, 93, 123, 125, 12288, 12288,
		2, 0, 45, 45, 60, 62, 1, 0, 34, 34, 3, 0, 9, 10, 13, 13, 32, 32, 4, 0,
		9, 10, 13, 13, 32, 32, 12288, 12288, 204, 0, 2, 1, 0, 0, 0, 0, 4, 1, 0,
		0, 0, 0, 6, 1, 0, 0, 0, 0, 8, 1, 0, 0, 0, 0, 10, 1, 0, 0, 0, 0, 12, 1,
		0, 0, 0, 0, 14, 1, 0, 0, 0, 0, 16, 1, 0, 0, 0, 0, 18, 1, 0, 0, 0, 0, 20,
		1, 0, 0, 0, 0, 22, 1, 0, 0, 0, 0, 24, 1, 0, 0, 0, 0, 30, 1, 0, 0, 0, 0,
		36, 1, 0, 0, 0, 0, 38, 1, 0, 0, 0, 0, 40, 1, 0, 0, 0, 0, 42, 1, 0, 0, 0,
		0, 44, 1, 0, 0, 0, 1, 54, 1, 0, 0, 0, 2, 56, 1, 0, 0, 0, 4, 58, 1, 0, 0,
		0, 6, 60, 1, 0, 0, 0, 8, 62, 1, 0, 0, 0, 10, 64, 1, 0, 0, 0, 12, 71, 1,
		0, 0, 0, 14, 77, 1, 0, 0, 0, 16, 84, 1, 0, 0, 0, 18, 86, 1, 0, 0, 0, 20,
		90, 1, 0, 0, 0, 22, 94, 1, 0, 0, 0, 24, 99, 1, 0, 0, 0, 26, 104, 1, 0,
		0, 0, 28, 107, 1, 0, 0, 0, 30, 109, 1, 0, 0, 0, 32, 118, 1, 0, 0, 0, 34,
		122, 1, 0, 0, 0, 36, 124, 1, 0, 0, 0, 38, 133, 1, 0, 0, 0, 40, 139, 1,
		0, 0, 0, 42, 147, 1, 0, 0, 0, 44, 158, 1, 0, 0, 0, 46, 164, 1, 0, 0, 0,
		48, 166, 1, 0, 0, 0, 50, 171, 1, 0, 0, 0, 52, 174, 1, 0, 0, 0, 54, 191,
		1, 0, 0, 0, 56, 57, 5, 40, 0, 0, 57, 3, 1, 0, 0, 0, 58, 59, 5, 41, 0, 0,
		59, 5, 1, 0, 0, 0, 60, 61, 5, 58, 0, 0, 61, 7, 1, 0, 0, 0, 62, 63, 5, 42,
		0, 0, 63, 9, 1, 0, 0, 0, 64, 65, 5, 34, 0, 0, 65, 11, 1, 0, 0, 0, 66, 67,
		5, 65, 0, 0, 67, 68, 5, 78, 0, 0, 68, 72, 5, 68, 0, 0, 69, 70, 5, 38, 0,
		0, 70, 72, 5, 38, 0, 0, 71, 66, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 72, 13,
		1, 0, 0, 0, 73, 74, 5, 79, 0, 0, 74, 78, 5, 82, 0, 0, 75, 76, 5, 124, 0,
		0, 76, 78, 5, 124, 0, 0, 77, 73, 1, 0, 0, 0, 77, 75, 1, 0, 0, 0, 78, 15,
		1, 0, 0, 0, 79, 80, 5, 78, 0, 0, 80, 81, 5, 79, 0, 0, 81, 85, 5, 84, 0,
		0, 82, 83, 7, 0, 0, 0, 83, 85, 4, 7, 0, 0, 84, 79, 1, 0, 0, 0, 84, 82,
		1, 0, 0, 0, 85, 17, 1, 0, 0, 0, 86, 87, 5, 60, 0, 0, 87, 88, 1, 0, 0, 0,
		88, 89, 6, 8, 0, 0, 89, 19, 1, 0, 0, 0, 90, 91, 5, 62, 0, 0, 91, 92, 1,
		0, 0, 0, 92, 93, 6, 9, 0, 0, 93, 21, 1, 0, 0, 0, 94, 95, 5, 60, 0, 0, 95,
		96, 5, 61, 0, 0, 96, 97, 1, 0, 0, 0, 97, 98, 6, 10, 0, 0, 98, 23, 1, 0,
		0, 0, 99, 100, 5, 62, 0, 0, 100, 101, 5, 61, 0, 0, 101, 102, 1, 0, 0, 0,
		102, 103, 6, 11, 0, 0, 103, 25, 1, 0, 0, 0, 104, 105, 5, 92, 0, 0, 105,
		106, 9, 0, 0, 0, 106, 27, 1, 0, 0, 0, 107, 108, 8, 1, 0, 0, 108, 29, 1,
		0, 0, 0, 109, 113, 3, 32, 15, 0, 110, 112, 3, 34, 16, 0, 111, 110, 1, 0,
		0, 0, 112, 115, 1, 0, 0, 0, 113, 111, 1, 0, 0, 0, 113, 114, 1, 0, 0, 0,
		114, 31, 1, 0, 0, 0, 115, 113, 1, 0, 0, 0, 116, 119, 3, 28, 13, 0, 117,
		119, 3, 26, 12, 0, 118, 116, 1, 0, 0, 0, 118, 117, 1, 0, 0, 0, 119, 33,
		1, 0, 0, 0, 120, 123, 3, 32, 15, 0, 121, 123, 7, 2, 0, 0, 122, 120, 1,
		0, 0, 0, 122, 121, 1, 0, 0, 0, 123, 35, 1, 0, 0, 0, 124, 128, 3, 32, 15,
		0, 125, 127, 3, 34, 16, 0, 126, 125, 1, 0, 0, 0, 127, 130, 1, 0, 0, 0,
		128, 126, 1, 0, 0, 0, 128, 129, 1, 0, 0, 0, 129, 131, 1, 0, 0, 0, 130,
		128, 1, 0, 0, 0, 131, 132, 3, 8, 3, 0, 132, 37, 1, 0, 0, 0, 133, 135, 3,
		8, 3, 0, 134, 136, 3, 34, 16, 0, 135, 134, 1, 0, 0, 0, 136, 137, 1, 0,
		0, 0, 137, 135, 1, 0, 0, 0, 137, 138, 1, 0, 0, 0, 138, 39, 1, 0, 0, 0,
		139, 141, 3, 8, 3, 0, 140, 142, 3, 34, 16, 0, 141, 140, 1, 0, 0, 0, 142,
		143, 1, 0, 0, 0, 143, 141, 1, 0, 0, 0, 143, 144, 1, 0, 0, 0, 144, 145,
		1, 0, 0, 0, 145, 146, 3, 8, 3, 0, 146, 41, 1, 0, 0, 0, 147, 152, 3, 10,
		4, 0, 148, 151, 3, 26, 12, 0, 149, 151, 8, 3, 0, 0, 150, 148, 1, 0, 0,
		0, 150, 149, 1, 0, 0, 0, 151, 154, 1, 0, 0, 0, 152, 153, 1, 0, 0, 0, 152,
		150, 1, 0, 0, 0, 153, 155, 1, 0, 0, 0, 154, 152, 1, 0, 0, 0, 155, 156,
		3, 10, 4, 0, 156, 43, 1, 0, 0, 0, 157, 159, 7, 4, 0, 0, 158, 157, 1, 0,
		0, 0, 159, 160, 1, 0, 0, 0, 160, 158, 1, 0, 0, 0, 160, 161, 1, 0, 0, 0,
		161, 162, 1, 0, 0, 0, 162, 163, 6, 21, 1, 0, 163, 45, 1, 0, 0, 0, 164,
		165, 8, 5, 0, 0, 165, 47, 1, 0, 0, 0, 166, 169, 3, 52, 25, 0, 167, 168,
		5, 69, 0, 0, 168, 170, 3, 52, 25, 0, 169, 167, 1, 0, 0, 0, 169, 170, 1,
		0, 0, 0, 170, 49, 1, 0, 0, 0, 171, 172, 2, 48, 57, 0, 172, 51, 1, 0, 0,
		0, 173, 175, 5, 45, 0, 0, 174, 173, 1, 0, 0, 0, 174, 175, 1, 0, 0, 0, 175,
		177, 1, 0, 0, 0, 176, 178, 3, 50, 24, 0, 177, 176, 1, 0, 0, 0, 178, 179,
		1, 0, 0, 0, 179, 177, 1, 0, 0, 0, 179, 180, 1, 0, 0, 0, 180, 187, 1, 0,
		0, 0, 181, 183, 5, 46, 0, 0, 182, 184, 3, 50, 24, 0, 183, 182, 1, 0, 0,
		0, 184, 185, 1, 0, 0, 0, 185, 183, 1, 0, 0, 0, 185, 186, 1, 0, 0, 0, 186,
		188, 1, 0, 0, 0, 187, 181, 1, 0, 0, 0, 187, 188, 1, 0, 0, 0, 188, 53, 1,
		0, 0, 0, 189, 192, 3, 48, 23, 0, 190, 192, 3, 30, 14, 0, 191, 189, 1, 0,
		0, 0, 191, 190, 1, 0, 0, 0, 192, 193, 1, 0, 0, 0, 193, 194, 6, 26, 2, 0,
		194, 195, 6, 26, 3, 0, 195, 55, 1, 0, 0, 0, 20, 0, 1, 71, 77, 84, 113,
		118, 122, 128, 137, 143, 150, 152, 160, 169, 174, 179, 185, 187, 191, 4,
		5, 1, 0, 6, 0, 0, 7, 13, 0, 4, 0, 0,
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

// QueryLexerInit initializes any static state used to implement QueryLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewQueryLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func QueryLexerInit() {
	staticData := &QueryLexerLexerStaticData
	staticData.once.Do(querylexerLexerInit)
}

// NewQueryLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewQueryLexer(input antlr.CharStream) *QueryLexer {
	QueryLexerInit()
	l := new(QueryLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &QueryLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "QueryLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// QueryLexer tokens.
const (
	QueryLexerLPAREN        = 1
	QueryLexerRPAREN        = 2
	QueryLexerCOLON         = 3
	QueryLexerSTAR          = 4
	QueryLexerDQUOTE        = 5
	QueryLexerAND           = 6
	QueryLexerOR            = 7
	QueryLexerNOT           = 8
	QueryLexerLT            = 9
	QueryLexerGT            = 10
	QueryLexerLT_EQ         = 11
	QueryLexerGT_EQ         = 12
	QueryLexerTERM          = 13
	QueryLexerTERM_PREFIX   = 14
	QueryLexerTERM_SUFFIX   = 15
	QueryLexerTERM_CONTAINS = 16
	QueryLexerPHRASE        = 17
	QueryLexerWS            = 18
)

// QueryLexerCOMPARISON is the QueryLexer mode.
const QueryLexerCOMPARISON = 1

func (l *QueryLexer) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 7:
		return l.NOT_Sempred(localctx, predIndex)

	default:
		panic("No registered predicate for: " + fmt.Sprint(ruleIndex))
	}
}

func (p *QueryLexer) NOT_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.GetInputStream().LA(1) != ' ' && p.GetInputStream().LA(1) != '\t'

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
